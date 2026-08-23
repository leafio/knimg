package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func testFileSize(t *testing.T, dir, name string) int64 {
	t.Helper()
	fi, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return fi.Size()
}

func makeTestPNG(t *testing.T, dir, name string, size int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)), 255})
		}
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func getJSON(t *testing.T, url string) map[string]interface{} {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func postJSON(t *testing.T, url string, payload interface{}) map[string]interface{} {
	t.Helper()
	data, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestAPIIntegration(t *testing.T) {
	workDir := t.TempDir()
	makeTestPNG(t, workDir, "photo.png", 64)
	makeTestPNG(t, workDir, "big.png", 500)
	if err := os.WriteFile(filepath.Join(workDir, "notes.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	mux, _ := InitServer(false)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("ListFiles", func(t *testing.T) {
		body := getJSON(t, ts.URL+"/api/files?work_dir="+workDir)
		if body["success"] != true {
			t.Fatalf("expected success, got %v", body)
		}
		if body["total"].(float64) != 3 {
			t.Fatalf("expected 3 files, got %v", body["total"])
		}
	})

	t.Run("FilterByType", func(t *testing.T) {
		body := getJSON(t, ts.URL+"/api/files?work_dir="+workDir+"&file_type=image")
		if body["total"].(float64) != 2 {
			t.Fatalf("expected 2 images, got %v", body["total"])
		}
	})

	t.Run("BrowseDirectory", func(t *testing.T) {
		body := getJSON(t, ts.URL+"/api/directory/browse?path="+workDir)
		if body["success"] != true {
			t.Fatalf("browse failed: %v", body)
		}
	})

	t.Run("CompressWithoutOverwriteKeepsOriginal", func(t *testing.T) {
		origPath := filepath.Join(workDir, "photo.png")
		fi, err := os.Stat(origPath)
		if err != nil {
			t.Fatal(err)
		}

		body := postJSON(t, ts.URL+"/api/compress", map[string]interface{}{
			"files":     []string{"photo.png"},
			"quality":   60,
			"work_dir":  workDir,
			"overwrite": false,
		})
		if body["success"] != true {
			t.Fatalf("compress failed: %v", body)
		}

		data := body["data"].(map[string]interface{})
		if data["failed_count"].(float64) != 0 {
			t.Fatalf("unexpected failures: %v", data["failed_files"])
		}

		fi2, err := os.Stat(origPath)
		if err != nil {
			t.Fatal(err)
		}
		if fi2.Size() != fi.Size() {
			t.Fatalf("original file was overwritten: %d -> %d", fi.Size(), fi2.Size())
		}

		compressedPath := filepath.Join(workDir, "compressed", "photo.png")
		if _, err := os.Stat(compressedPath); err != nil {
			t.Fatalf("compressed output missing: %v", err)
		}
	})

	t.Run("CompressOverwriteAtomic", func(t *testing.T) {
		before, err := os.Stat(filepath.Join(workDir, "big.png"))
		if err != nil {
			t.Fatal(err)
		}

		body := postJSON(t, ts.URL+"/api/compress", map[string]interface{}{
			"files":     []string{"big.png"},
			"quality":   50,
			"work_dir":  workDir,
			"overwrite": true,
		})
		if body["success"] != true {
			t.Fatalf("compress failed: %v", body)
		}
		data := body["data"].(map[string]interface{})
		if data["failed_count"].(float64) != 0 {
			t.Fatalf("unexpected failures: %v", data["failed_files"])
		}

		after, err := os.Stat(filepath.Join(workDir, "big.png"))
		if err != nil {
			t.Fatal(err)
		}
		if after.Size() >= before.Size() {
			t.Logf("warning: recompression did not shrink file (%d -> %d)", before.Size(), after.Size())
		}

		matches, _ := filepath.Glob(filepath.Join(workDir, "*.tmp"))
		if len(matches) != 0 {
			t.Fatalf("temp files left behind: %v", matches)
		}
	})

	t.Run("ServeImageThumbAndFull", func(t *testing.T) {
		// big.png 256px > 512B，应返回缩略图(jpeg)而非原文件
		resp, err := http.Get(ts.URL + "/api/file/image?work_dir=" + workDir + "&path=big.png")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("thumb status %d", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "image/jpeg" {
			t.Fatalf("expected thumbnail jpeg, got %s", ct)
		}
		thumbBytes, _ := io.ReadAll(resp.Body)
		if len(thumbBytes) == 0 || len(thumbBytes) >= int(testFileSize(t, workDir, "big.png")) {
			t.Fatalf("thumbnail not smaller than original: %d bytes", len(thumbBytes))
		}

		// full=1 应返回原始 PNG
		resp2, err := http.Get(ts.URL + "/api/file/image?work_dir=" + workDir + "&path=big.png&full=1")
		if err != nil {
			t.Fatal(err)
		}
		defer resp2.Body.Close()
		if ct := resp2.Header.Get("Content-Type"); ct != "image/png" {
			t.Fatalf("expected original png, got %s", ct)
		}
	})

	t.Run("ServeImageRejectsPathEscape", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/file/image?work_dir=" + workDir + "&path=../../etc/passwd")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 400 {
			t.Fatalf("expected error for path escape, got %d", resp.StatusCode)
		}
	})

	t.Run("ExportCSV", func(t *testing.T) {
		exportDir := t.TempDir()
		body := getJSON(t, ts.URL+"/api/files/export?format=csv&export_dir="+exportDir+"&work_dir="+workDir)
		if body["success"] != true {
			t.Fatalf("export failed: %v", body)
		}
		if _, err := os.Stat(filepath.Join(exportDir, body["file_path"].(string))); err != nil {
			t.Fatalf("exported file missing: %v", err)
		}
	})
}
