package handlers

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	_ "image/gif" // 支持 gif
	"knimg/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
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
func (h *CompressHandler) CompressFiles(c *gin.Context) {
	var req models.CompressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	if len(req.Files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
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
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "工作目录不存在: " + workDir,
		})
		return
	}

	// 设置输出目录
	outputDir := req.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(workDir, "compressed")
	}

	// 创建输出目录（如果不覆盖原文件）
	if !req.Overwrite {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "无法创建输出目录",
			})
			return
		}
	}

	// 设置响应头为流式
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("X-Content-Type-Options", "nosniff")

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
			progress := (i + 1) * 100 / fileCount
			fmt.Fprintf(c.Writer, "progress:%d\n", progress)
			c.Writer.Flush()
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
		progress := (i + 1) * 100 / fileCount
		fmt.Fprintf(c.Writer, "progress:%d\n", progress)
		c.Writer.Flush()
	}

	// 构建响应
	response := gin.H{
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

	// 发送最终响应
	fmt.Fprintf(c.Writer, "data:%s\n", toJSON(response))
	c.Writer.Flush()
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
	outputPath := filepath.Join(outputDir, filename)

	// 如果是覆盖原文件，先删除原文件
	if overwrite && inputPath == outputPath {
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
	var newSize int64
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
		if err == nil {
			fileInfo, _ := os.Stat(outputPath)
			newSize = fileInfo.Size()
		}
	case "png":
		// PNG 使用最佳压缩，保持无损
		err = png.Encode(outputFile, img)
		if err == nil {
			fileInfo, _ := os.Stat(outputPath)
			newSize = fileInfo.Size()
		}
	case "gif":
		// GIF 转为 JPEG（GIF 压缩效果差）
		newFilename := strings.TrimSuffix(filename, ext) + ".jpg"
		outputPath = filepath.Join(outputDir, newFilename)
		outputFile.Close()
		
		newFile, err := os.Create(outputPath)
		if err != nil {
			return 0, err
		}
		defer newFile.Close()
		
		err = jpeg.Encode(newFile, img, &jpeg.Options{Quality: quality})
		if err == nil {
			fileInfo, _ := os.Stat(outputPath)
			newSize = fileInfo.Size()
		}
	default:
		// WebP 等其他格式转为 JPEG
		newFilename := strings.TrimSuffix(filename, ext) + ".jpg"
		outputPath = filepath.Join(outputDir, newFilename)
		outputFile.Close()
		
		newFile, err := os.Create(outputPath)
		if err != nil {
			return 0, err
		}
		defer newFile.Close()
		
		err = jpeg.Encode(newFile, img, &jpeg.Options{Quality: quality})
		if err == nil {
			fileInfo, _ := os.Stat(outputPath)
			newSize = fileInfo.Size()
		}
	}

	if err != nil {
		return 0, err
	}

	return newSize, nil
}

// GetCompressionStats 获取压缩统计信息
func (h *CompressHandler) GetCompressionStats(c *gin.Context) {
	// 支持自定义压缩目录
	compressedDir := c.Query("compressed_dir")
	if compressedDir == "" {
		compressedDir = filepath.Join(h.BaseDir, "compressed")
	}

	// 检查压缩目录是否存在
	if _, err := os.Stat(compressedDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"total_files": 0,
				"total_size":  0,
			},
		})
		return
	}

	// 扫描压缩后的文件
	var totalFiles int
	var totalSize int64

	err := filepath.Walk(compressedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_files": totalFiles,
			"total_size":  totalSize,
		},
	})
}
