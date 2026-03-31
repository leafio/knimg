package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"knimg/models"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx/v3"
)

// FileHandler 文件处理处理器
type FileHandler struct {
	BaseDir string
}

// NewFileHandler 创建文件处理器
func NewFileHandler(baseDir string) *FileHandler {
	return &FileHandler{
		BaseDir: baseDir,
	}
}

// ListFiles 获取文件列表（支持筛选和排序）
func (h *FileHandler) ListFiles(c *gin.Context) {
	var req models.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 支持自定义工作目录
	workDir := c.Query("work_dir")
	if workDir == "" {
		workDir = h.BaseDir
	}

	// 验证目录存在
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "目录不存在: " + workDir,
		})
		return
	}

	files, err := h.scanFilesWithFilter(workDir, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 计算统计信息
	stats := h.calculateStats(files)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      files,
		"total":     len(files),
		"stats":     stats,
		"work_dir":  workDir,
	})
}

// scanFilesWithFilter 扫描目录中的文件（带筛选）
func (h *FileHandler) scanFilesWithFilter(workDir string, req *models.FileListRequest) ([]models.FileInfo, error) {
	var files []models.FileInfo

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(workDir, path)
			fileType := getFileType(info.Name())
			fileExt := strings.ToLower(filepath.Ext(info.Name()))
			
			file := models.FileInfo{
				Name:       info.Name(),
				Path:       relPath,
				Size:       info.Size(),
				Type:       fileType,
				Ext:        fileExt,
				ModTime:    info.ModTime().Format("2006-01-02 15:04:05"),
				Compressed: false,
			}

			// 应用筛选条件
			if h.matchesFilter(file, req) {
				files = append(files, file)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 应用排序
	h.sortFiles(files, req)

	return files, nil
}

// matchesFilter 检查文件是否匹配筛选条件
func (h *FileHandler) matchesFilter(file models.FileInfo, req *models.FileListRequest) bool {
	// 名称模糊搜索
	if req.Search != "" && !strings.Contains(strings.ToLower(file.Name), strings.ToLower(req.Search)) {
		return false
	}

	// 文件类型筛选
	if req.FileType != "" && req.FileType != "all" && file.Type != req.FileType {
		return false
	}

	// 自定义扩展名筛选
	if req.FileExt != "" {
		extList := strings.Split(req.FileExt, ",")
		matched := false
		for _, ext := range extList {
			ext = strings.TrimSpace(ext)
			ext = strings.ToLower(ext)
			if ext == "" {
				continue
			}
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			if file.Ext == ext {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 大小范围筛选
	if req.MinSize > 0 && file.Size < req.MinSize {
		return false
	}
	if req.MaxSize > 0 && file.Size > req.MaxSize {
		return false
	}

	return true
}

// sortFiles 对文件列表进行排序
func (h *FileHandler) sortFiles(files []models.FileInfo, req *models.FileListRequest) {
	if req.SortBy == "" {
		req.SortBy = "name" // 默认按名称排序
	}
	if req.SortOrder == "" {
		req.SortOrder = "asc" // 默认升序
	}

	sort.Slice(files, func(i, j int) bool {
		var result bool
		switch req.SortBy {
		case "name":
			result = files[i].Name < files[j].Name
		case "size":
			result = files[i].Size < files[j].Size
		case "type":
			result = files[i].Type < files[j].Type
		case "time":
			result = files[i].ModTime < files[j].ModTime
		default:
			result = files[i].Name < files[j].Name
		}

		if req.SortOrder == "desc" {
			return !result
		}
		return result
	})
}

// calculateStats 计算文件统计信息
func (h *FileHandler) calculateStats(files []models.FileInfo) gin.H {
	var totalSize int64
	var imageCount, docCount, videoCount, otherCount int

	for _, file := range files {
		totalSize += file.Size
		switch file.Type {
		case "image":
			imageCount++
		case "document":
			docCount++
		case "video":
			videoCount++
		default:
			otherCount++
		}
	}

	return gin.H{
		"total_files":   len(files),
		"total_size":    totalSize,
		"image_count":   imageCount,
		"doc_count":     docCount,
		"video_count":   videoCount,
		"other_count":   otherCount,
	}
}

// getFileType 根据扩展名获取文件类型
func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return "image"
	case ".pdf", ".doc", ".docx", ".txt", ".xls", ".xlsx":
		return "document"
	case ".mp4", ".avi", ".mov", ".mkv", ".flv":
		return "video"
	default:
		return "other"
	}
}

// ExportFiles 导出文件列表
func (h *FileHandler) ExportFiles(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "excel"
	}

	// 支持自定义导出目录
	exportDir := c.Query("export_dir")
	if exportDir == "" {
		exportDir = h.BaseDir
	}

	// 创建导出目录
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "无法创建导出目录: " + err.Error(),
		})
		return
	}

	// 获取筛选参数
	var req models.FileListRequest
	c.ShouldBindQuery(&req)

	// 支持自定义工作目录
	workDir := c.Query("work_dir")
	if workDir == "" {
		workDir = h.BaseDir
	}

	// 验证目录存在
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "目录不存在: " + workDir,
		})
		return
	}

	files, err := h.scanFilesWithFilter(workDir, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var filePath string
	switch format {
	case "excel":
		filePath, err = h.exportToExcel(files, exportDir)
	case "csv":
		filePath, err = h.exportToCSV(files, exportDir)
	case "json":
		filePath, err = h.exportToJSON(files, exportDir)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的导出格式",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"file_path":  filePath,
		"export_dir": exportDir,
	})
}

// exportToExcel 导出为 Excel
func (h *FileHandler) exportToExcel(files []models.FileInfo, exportDir string) (string, error) {
	xlFile := xlsx.NewFile()
	sheet, _ := xlFile.AddSheet("文件列表")

	// 添加表头
	headerRow := sheet.AddRow()
	headers := []string{"文件名", "路径", "大小 (字节)", "类型", "扩展名", "修改时间"}
	for _, header := range headers {
		cell := headerRow.AddCell()
		cell.Value = header
	}

	// 填充数据
	for _, file := range files {
		row := sheet.AddRow()
		
		cell1 := row.AddCell()
		cell1.Value = file.Name
		
		cell2 := row.AddCell()
		cell2.Value = file.Path
		
		cell3 := row.AddCell()
		cell3.Value = fmt.Sprintf("%d", file.Size)
		
		cell4 := row.AddCell()
		cell4.Value = file.Type
		
		cell5 := row.AddCell()
		cell5.Value = file.Ext
		
		cell6 := row.AddCell()
		cell6.Value = file.ModTime
	}

	// 保存文件
	filename := fmt.Sprintf("files_%s.xlsx", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(exportDir, filename)
	if err := xlFile.Save(filepath); err != nil {
		return "", err
	}

	return filename, nil
}

// exportToCSV 导出为 CSV
func (h *FileHandler) exportToCSV(files []models.FileInfo, exportDir string) (string, error) {
	filename := fmt.Sprintf("files_%s.csv", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(exportDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"文件名", "路径", "大小 (字节)", "类型", "扩展名", "修改时间"}
	if err := writer.Write(headers); err != nil {
		return "", err
	}

	// 写入数据
	for _, file := range files {
		record := []string{
			file.Name,
			file.Path,
			fmt.Sprintf("%d", file.Size),
			file.Type,
			file.Ext,
			file.ModTime,
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	return filename, nil
}

// exportToJSON 导出为 JSON
func (h *FileHandler) exportToJSON(files []models.FileInfo, exportDir string) (string, error) {
	filename := fmt.Sprintf("files_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(exportDir, filename)

	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", err
	}

	return filename, nil
}

// GetHomeDirectory 获取用户主目录
func (h *FileHandler) GetHomeDirectory(c *gin.Context) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/"
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"path":    homeDir,
	})
}

// BrowseDirectory 浏览目录
func (h *FileHandler) BrowseDirectory(c *gin.Context) {
	dirPath := c.Query("path")
	if dirPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			dirPath = "/"
		} else {
			dirPath = homeDir
		}
	}

	// 验证目录存在
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "目录不存在",
		})
		return
	}

	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不是目录",
		})
		return
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var directories []string
	// 添加父目录
	if dirPath != "/" {
		parentDir := filepath.Dir(dirPath)
		if parentDir == dirPath {
			parentDir = "/"
		}
		directories = append(directories, "..")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"current_path": dirPath,
		"directories":  directories,
	})
}
