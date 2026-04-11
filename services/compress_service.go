package services

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	_ "image/gif"       // 支持 gif
	_ "golang.org/x/image/webp" // 支持 webp
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
	Success      bool
	Message      string
	OutputPath   string
	OrigSize     int64
	NewSize      int64
	FailedFiles  []string
	FailedCount  int
}

// GetOutputDir 获取输出目录
func (s *CompressService) GetOutputDir(workDir, outputDir string, overwrite bool) string {
	if outputDir != "" {
		return outputDir
	}

	if overwrite {
		return workDir
	}

	return workDir
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

// CompressImage 压缩单个图片文件
func (s *CompressService) CompressImage(inputPath, outputDir string, quality int, overwrite bool) (int64, error) {
	// 打开输入文件
	file, err := os.Open(inputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// 解码图片
	img, format, err := image.Decode(file)
	if err != nil {
		return 0, fmt.Errorf("无法解码图片: %v", err)
	}

	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return 0, err
	}

	// 获取原始文件大小
	_, err = file.Stat()
	if err != nil {
		return 0, err
	}

	// 确定输出路径
	var outputPath string
	if overwrite {
		outputPath = inputPath
	} else {
		filename := filepath.Base(inputPath)
		outputPath = filepath.Join(outputDir, filename)
	}

	// 创建输出目录（如果需要）
	if !overwrite {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return 0, err
		}
	}

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer outFile.Close()

	// 根据图片格式进行压缩
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(outFile, img)
	default:
		// 对于其他格式，尝试转换为 JPEG
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return 0, fmt.Errorf("压缩失败: %v", err)
	}

	// 获取压缩后的文件大小
	outFileInfo, err := outFile.Stat()
	if err != nil {
		return 0, err
	}

	return outFileInfo.Size(), nil
}

// CompressFiles 批量压缩图片
func (s *CompressService) CompressFiles(files []string, quality int, workDir, outputDir string, overwrite bool) CompressResult {
	result := CompressResult{
		Success:     true,
		FailedFiles: []string{},
	}

	// 智能质量选择
	if quality <= 0 || quality > 100 {
		quality = 80 // 默认质量
	}

	// 支持自定义工作目录
	if workDir == "" {
		workDir = s.BaseDir
	}

	// 验证工作目录存在
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		result.Success = false
		result.Message = "工作目录不存在: " + workDir
		return result
	}

	// 设置输出目录
	outputDir = s.GetOutputDir(workDir, outputDir, overwrite)

	// 创建输出目录（如果不覆盖原文件）
	if !overwrite {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			result.Success = false
			result.Message = "无法创建输出目录"
			return result
		}
	}

	var totalOrigSize, totalNewSize int64
	var successCount int

	for _, filePath := range files {
		fullPath := filepath.Join(workDir, filePath)

		// 检查文件是否存在
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			result.FailedFiles = append(result.FailedFiles, filePath+" (文件不存在)")
			continue
		}

		// 智能压缩策略：根据原文件大小调整质量
		adjustedQuality := s.GetSmartQuality(fileInfo.Size(), quality)

		// 压缩文件
		outputPath := outputDir
		if overwrite {
			outputPath = filepath.Dir(fullPath)
		}
		newSize, err := s.CompressImage(fullPath, outputPath, adjustedQuality, overwrite)
		if err != nil {
			result.FailedFiles = append(result.FailedFiles, filePath+" ("+err.Error()+")")
		} else {
			totalOrigSize += fileInfo.Size()
			totalNewSize += newSize
			successCount++
		}
	}

	result.Message = fmt.Sprintf("成功压缩 %d 个图片", successCount)
	result.OutputPath = outputDir
	result.OrigSize = totalOrigSize
	result.NewSize = totalNewSize
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
