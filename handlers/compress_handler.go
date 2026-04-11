package handlers

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	_ "image/gif"       // 支持 gif
	_ "golang.org/x/image/webp" // 支持 webp
	"knimg/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CompressHandler 图片压缩处理器
type CompressHandler struct {
	BaseDir string
}

// NewCompressHandler 创建压缩处理器
func NewCompressHandler(baseDir string) *CompressHandler {
	return &CompressHandler{
		BaseDir: baseDir,
	}
}

// CompressFiles 批量压缩图片（支持流式进度）
func (h *CompressHandler) CompressFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	if len(req.Files) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未选择要压缩的文件",
		})
		return
	}

	// 智能质量选择
	quality := req.Quality
	if quality <= 0 || quality > 100 {
		quality = 80 // 默认质量
	}

	// 支持自定义工作目录
	workDir := req.WorkDir
	if workDir == "" {
		workDir = h.BaseDir
	}

	// 验证工作目录存在
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "工作目录不存在: " + workDir,
		})
		return
	}

	// 设置输出目录
	outputDir := h.getOutputDir(workDir, req.OutputDir, req.Overwrite)

	// 创建输出目录（如果不覆盖原文件）
	if !req.Overwrite {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "无法创建输出目录",
			})
			return
		}
	}

	// 设置响应头为流式
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	var totalOrigSize, totalNewSize int64
	var successCount int
	var failedFiles []string
	fileCount := len(req.Files)

	for i, filePath := range req.Files {
		fullPath := filepath.Join(workDir, filePath)

		// 检查文件是否存在
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			failedFiles = append(failedFiles, filePath+" (文件不存在)")
			// 更新进度
			h.updateProgress(w, i+1, fileCount)
			continue
		}

		// 智能压缩策略：根据原文件大小调整质量
		adjustedQuality := h.getSmartQuality(fileInfo.Size(), quality)

		// 压缩文件
		outputPath := outputDir
		if req.Overwrite {
			outputPath = filepath.Dir(fullPath)
		}
		newSize, err := h.compressImage(fullPath, outputPath, adjustedQuality, req.Overwrite)
		if err != nil {
			failedFiles = append(failedFiles, filePath+" ("+err.Error()+")")
		} else {
			totalOrigSize += fileInfo.Size()
			totalNewSize += newSize
			successCount++
		}

		// 更新进度
		h.updateProgress(w, i+1, fileCount)
	}

	// 构建响应
	response := h.buildResponse(successCount, totalOrigSize, totalNewSize, outputDir, workDir, failedFiles)

	// 发送最终响应
	fmt.Fprintf(w, "data:%s\n", toJSON(response))
	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

// getOutputDir 获取输出目录
func (h *CompressHandler) getOutputDir(workDir, outputDir string, overwrite bool) string {
	if outputDir != "" {
		return outputDir
	}

	if overwrite {
		return workDir
	}

	return filepath.Join(workDir, "compressed")
}

// updateProgress 更新压缩进度
func (h *CompressHandler) updateProgress(w http.ResponseWriter, current, total int) {
	progress := current * 100 / total
	fmt.Fprintf(w, "progress:%d\n", progress)
	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

// buildResponse 构建响应对象
func (h *CompressHandler) buildResponse(successCount int, totalOrigSize, totalNewSize int64, outputDir, workDir string, failedFiles []string) map[string]interface{} {
	response := map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("成功压缩 %d 个图片", successCount),
		"output_path": outputDir,
		"orig_size":   totalOrigSize,
		"new_size":    totalNewSize,
		"ratio":       fmt.Sprintf("%.2f%%", float64(totalNewSize)/float64(totalOrigSize)*100),
		"work_dir":    workDir,
	}

	if len(failedFiles) > 0 {
		response["failed_files"] = failedFiles
		response["failed_count"] = len(failedFiles)
	}

	return response
}

// toJSON 转换为 JSON 字符串
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// getSmartQuality 智能质量选择策略
func (h *CompressHandler) getSmartQuality(fileSize int64, baseQuality int) int {
	// 根据文件大小调整压缩质量
	// 小文件 (< 100KB): 使用较高质量，避免过度压缩
	// 中等文件 (100KB - 1MB): 使用基础质量
	// 大文件 (> 1MB): 可以适当降低质量以获得更好的压缩率

	const (
		KB = 1024
		MB = 1024 * KB
	)

	switch {
	case fileSize < 100*KB:
		// 小文件：提高质量，最小85
		if baseQuality < 85 {
			return 85
		}
		return baseQuality
	case fileSize > 1*MB:
		// 大文件：可以适当降低质量，但不低于60
		adjusted := baseQuality - 10
		if adjusted < 60 {
			return 60
		}
		return adjusted
	default:
		// 中等文件：使用基础质量
		return baseQuality
	}
}

// compressImage 压缩单个图片（支持覆盖原文件）
func (h *CompressHandler) compressImage(inputPath, outputDir string, quality int, overwrite bool) (int64, error) {
	// 打开图片文件
	file, err := os.Open(inputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// 解码图片
	img, format, err := image.Decode(file)
	if err != nil {
		return 0, err
	}

	// 创建输出文件
	filename := filepath.Base(inputPath)
	ext := strings.ToLower(filepath.Ext(filename))

	// 智能格式选择：WebP 统一转为 JPEG 以获得更好兼容性
	// PNG 保持 PNG（无损），其他转为 JPEG
	outputPath := inputPath
	if !overwrite {
		outputPath = filepath.Join(outputDir, filename)
	}

	// 如果是覆盖原文件，先删除原文件
	if overwrite {
		if err := os.Remove(inputPath); err != nil {
			return 0, err
		}
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer outputFile.Close()

	// 根据格式保存图片
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	case "png":
		// PNG 使用最佳压缩，保持无损
		err = png.Encode(outputFile, img)
	case "gif", "webp":
		// GIF/WebP 转为 JPEG（GIF 压缩效果差，WebP 兼容性不好）
		newFilename := strings.TrimSuffix(filename, ext) + ".jpg"
		if overwrite {
			outputPath = strings.TrimSuffix(inputPath, ext) + ".jpg"
		} else {
			outputPath = filepath.Join(outputDir, newFilename)
			outputFile.Close()

			outputFile, err = os.Create(outputPath)
			if err != nil {
				return 0, err
			}
			defer outputFile.Close()
		}

		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	default:
		// 其他格式转为 JPEG
		newFilename := strings.TrimSuffix(filename, ext) + ".jpg"
		if overwrite {
			outputPath = strings.TrimSuffix(inputPath, ext) + ".jpg"
		} else {
			outputPath = filepath.Join(outputDir, newFilename)
			outputFile.Close()

			outputFile, err = os.Create(outputPath)
			if err != nil {
				return 0, err
			}
			defer outputFile.Close()
		}

		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return 0, err
	}

	fileInfo, _ := os.Stat(outputPath)
	return fileInfo.Size(), nil
}

// GetCompressionStats 获取压缩统计信息
func (h *CompressHandler) GetCompressionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 支持自定义压缩目录
	compressedDir := r.URL.Query().Get("compressed_dir")
	if compressedDir == "" {
		compressedDir = filepath.Join(h.BaseDir, "compressed")
	}

	// 检查压缩目录是否存在
	if _, err := os.Stat(compressedDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"total_files": 0,
				"total_size":  0,
			},
		})
		return
	}

	// 扫描压缩后的文件
	totalFiles, totalSize := h.scanDirectory(compressedDir)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"total_files": totalFiles,
			"total_size":  totalSize,
		},
	})
}

// scanDirectory 扫描目录统计文件数量和大小
func (h *CompressHandler) scanDirectory(dirPath string) (int, int64) {
	var totalFiles int
	var totalSize int64

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}
		return nil
	})

	return totalFiles, totalSize
}
