package services

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	_ "golang.org/x/image/bmp"  // 支持 bmp
	_ "golang.org/x/image/webp" // 支持 webp
	_ "image/gif"               // 支持 gif
)

// CompressService 图片压缩服务
type CompressService struct {
	BaseDir string
}

// NewCompressService 创建压缩服务实例
func NewCompressService(baseDir string) *CompressService {
	return &CompressService{
		BaseDir: baseDir,
	}
}

// CompressResult 压缩结果
type CompressResult struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	OutputPath  string   `json:"output_path"`
	OrigSize    int64    `json:"orig_size"`
	NewSize     int64    `json:"new_size"`
	FailedFiles []string `json:"failed_files,omitempty"`
	FailedCount int      `json:"failed_count"`
}

// GetOutputDir 获取输出目录
func (s *CompressService) GetOutputDir(workDir, outputDir string, overwrite bool) string {
	if overwrite {
		return workDir
	}
	if outputDir != "" && filepath.Clean(outputDir) != filepath.Clean(workDir) {
		return outputDir
	}
	return filepath.Join(workDir, "compressed")
}

// GetSmartQuality 根据文件大小获取智能压缩质量
func (s *CompressService) GetSmartQuality(fileSize int64, baseQuality int) int {
	quality := baseQuality

	// 根据文件大小调整质量
	switch {
	case fileSize > 10*1024*1024: // >10MB
		quality = quality - 15
	case fileSize > 5*1024*1024: // >5MB
		quality = quality - 10
	case fileSize > 2*1024*1024: // >2MB
		quality = quality - 5
	case fileSize < 100*1024: // <100KB
		quality = quality + 10
	}

	// 确保质量在合理范围内
	if quality < 30 {
		quality = 30
	} else if quality > 100 {
		quality = 100
	}

	return quality
}

// resolveOutputFormat 确定输出格式，返回标准化格式名
func resolveOutputFormat(target, source string) string {
	switch strings.ToLower(strings.TrimPrefix(target, ".")) {
	case "jpeg", "jpg":
		return "jpeg"
	case "png":
		return "png"
	default:
		if source == "png" {
			return "png"
		}
		return "jpeg"
	}
}

func extForFormat(format string) string {
	if format == "png" {
		return ".png"
	}
	return ".jpg"
}

// CompressImage 压缩单个图片文件
func (s *CompressService) CompressImage(inputPath, outputDir string, quality int, overwrite bool, targetFormat string) (int64, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	img, srcFormat, err := image.Decode(file)
	if err != nil {
		return 0, fmt.Errorf("无法解码图片: %v", err)
	}

	outFormat := resolveOutputFormat(targetFormat, srcFormat)

	var outputPath string
	if overwrite && outFormat == srcFormat {
		outputPath = inputPath
	} else if overwrite {
		filename := filepath.Base(inputPath)
		outputPath = filepath.Join(filepath.Dir(inputPath),
			strings.TrimSuffix(filename, filepath.Ext(filename))+extForFormat(outFormat))
	} else {
		filename := filepath.Base(inputPath)
		ext := extForFormat(outFormat)
		base := strings.TrimSuffix(filename, filepath.Ext(filename))
		if filepath.Clean(outputDir) == filepath.Clean(filepath.Dir(inputPath)) {
			base += "_compressed"
		}
		outputPath = filepath.Join(outputDir, base+ext)
	}

	var tmpPath string
	var outFile *os.File
	if overwrite || outFormat != srcFormat {
		tmp, err := os.CreateTemp(filepath.Dir(outputPath), filepath.Base(outputPath)+".*.tmp")
		if err != nil {
			return 0, err
		}
		tmpPath = tmp.Name()
		tmp.Close()
		outFile, err = os.Create(tmpPath)
		if err != nil {
			os.Remove(tmpPath)
			return 0, err
		}
	} else {
		outFile, err = os.Create(outputPath)
		if err != nil {
			return 0, err
		}
	}

	if outFormat == "png" {
		err = png.Encode(outFile, img)
	} else {
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	}
	closeErr := outFile.Close()

	if err != nil || closeErr != nil {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
		if err != nil {
			return 0, fmt.Errorf("压缩失败: %v", err)
		}
		return 0, closeErr
	}

	if tmpPath != "" {
		if err := os.Rename(tmpPath, outputPath); err != nil {
			os.Remove(tmpPath)
			return 0, err
		}
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CompressFiles 批量压缩图片
func (s *CompressService) CompressFiles(files []string, quality int, workDir, outputDir string, overwrite bool, targetFormat string) CompressResult {
	result := CompressResult{
		Success:     true,
		FailedFiles: []string{},
	}

	if quality <= 0 || quality > 100 {
		quality = 80
	}

	if workDir == "" {
		workDir = s.BaseDir
	}

	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		result.Success = false
		result.Message = "工作目录不存在: " + workDir
		return result
	}

	outputDir = s.GetOutputDir(workDir, outputDir, overwrite)

	if !overwrite {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			result.Success = false
			result.Message = "无法创建输出目录"
			return result
		}
	}

	var mu sync.Mutex
	jobs := make(chan string)
	workers := runtime.NumCPU()
	if workers > 8 {
		workers = 8
	}
	if workers < 1 {
		workers = 1
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				fullPath := filepath.Join(workDir, filePath)

				fileInfo, err := os.Stat(fullPath)
				if err != nil {
					mu.Lock()
					result.FailedFiles = append(result.FailedFiles, filePath+" (文件不存在)")
					mu.Unlock()
					continue
				}

				adjustedQuality := s.GetSmartQuality(fileInfo.Size(), quality)

				targetDir := outputDir
				if overwrite {
					targetDir = filepath.Dir(fullPath)
				}
				newSize, err := s.CompressImage(fullPath, targetDir, adjustedQuality, overwrite, targetFormat)

				mu.Lock()
				if err != nil {
					result.FailedFiles = append(result.FailedFiles, filePath+" ("+err.Error()+")")
				} else {
					result.OrigSize += fileInfo.Size()
					result.NewSize += newSize
				}
				mu.Unlock()
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	wg.Wait()

	successCount := len(files) - len(result.FailedFiles)
	result.Message = fmt.Sprintf("成功压缩 %d 个图片", successCount)
	result.OutputPath = outputDir
	result.FailedCount = len(result.FailedFiles)

	return result
}

// GetCompressionStats 获取压缩统计信息
func (s *CompressService) GetCompressionStats(compressedDir string) (int, int64) {
	if compressedDir == "" {
		return 0, 0
	}

	// 检查压缩目录是否存在
	if _, err := os.Stat(compressedDir); os.IsNotExist(err) {
		return 0, 0
	}

	return s.ScanDirectory(compressedDir)
}

// ScanDirectory 扫描目录统计文件数量和大小
func (s *CompressService) ScanDirectory(dirPath string) (int, int64) {
	var totalFiles int
	var totalSize int64

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}
		return nil
	})

	return totalFiles, totalSize
}
