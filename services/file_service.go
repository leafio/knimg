package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"knimg/models"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tealeg/xlsx/v3"
)

// FileService 文件管理服务
type FileService struct {
	BaseDir string
}

// NewFileService 创建文件服务实例
func NewFileService(baseDir string) *FileService {
	return &FileService{
		BaseDir: baseDir,
	}
}

// GetWorkDir 获取工作目录，如果未指定则使用默认目录
func (s *FileService) GetWorkDir(queryDir string) string {
	workDir := queryDir
	if workDir == "" {
		workDir = s.BaseDir
	}

	// 验证目录存在
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return s.BaseDir
	}
	return workDir
}

// ScanFilesWithFilter 扫描目录中的文件（带筛选）
func (s *FileService) ScanFilesWithFilter(workDir string, req *models.FileListRequest) ([]models.FileInfo, error) {
	var files []models.FileInfo

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(workDir, path)
			fileType := s.GetFileType(info.Name())
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
			if s.MatchesFilter(file, req) {
				files = append(files, file)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 应用排序
	s.SortFiles(files, req)

	return files, nil
}

// GetFileType 根据文件名获取文件类型
func (s *FileService) GetFileType(filename string) string {
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

// MatchesFilter 检查文件是否匹配筛选条件
func (s *FileService) MatchesFilter(file models.FileInfo, req *models.FileListRequest) bool {
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

// SortFiles 对文件列表进行排序
func (s *FileService) SortFiles(files []models.FileInfo, req *models.FileListRequest) {
	if req.SortBy == "" {
		req.SortBy = "name" // 默认按名称排序
	}

	sort.Slice(files, func(i, j int) bool {
		a, b := files[i], files[j]
		switch req.SortBy {
		case "name":
			if req.SortOrder == "desc" {
				return a.Name > b.Name
			}
			return a.Name < b.Name
		case "size":
			if req.SortOrder == "desc" {
				return a.Size > b.Size
			}
			return a.Size < b.Size
		case "type":
			if req.SortOrder == "desc" {
				return a.Type > b.Type
			}
			return a.Type < b.Type
		case "time":
			if req.SortOrder == "desc" {
				return a.ModTime > b.ModTime
			}
			return a.ModTime < b.ModTime
		default:
			return a.Name < b.Name
		}
	})
}

// CalculateStats 计算文件统计信息
func (s *FileService) CalculateStats(files []models.FileInfo) map[string]interface{} {
	var imageCount, docCount, videoCount, otherCount int

	for _, file := range files {
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

	return map[string]interface{}{
		"image_count": imageCount,
		"doc_count":   docCount,
		"video_count": videoCount,
		"other_count": otherCount,
	}
}

// ExportToExcel 导出为 Excel
func (s *FileService) ExportToExcel(files []models.FileInfo, exportDir string) (string, error) {
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
		
		row.AddCell().SetValue(file.Name)
		row.AddCell().SetValue(file.Path)
		row.AddCell().SetValue(file.Size)
		row.AddCell().SetValue(file.Type)
		row.AddCell().SetValue(file.Ext)
		row.AddCell().SetValue(file.ModTime)
	}

	// 保存文件
	filename := fmt.Sprintf("files_%s.xlsx", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(exportDir, filename)
	if err := xlFile.Save(filePath); err != nil {
		return "", err
	}

	return filename, nil
}

// ExportToCSV 导出为 CSV
func (s *FileService) ExportToCSV(files []models.FileInfo, exportDir string) (string, error) {
	filename := fmt.Sprintf("files_%s.csv", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(exportDir, filename)

	file, err := os.Create(filePath)
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

// ExportToJSON 导出为 JSON
func (s *FileService) ExportToJSON(files []models.FileInfo, exportDir string) (string, error) {
	filename := fmt.Sprintf("files_%s.json", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(exportDir, filename)

	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return filename, nil
}

// GetHomeDirectory 获取用户主目录
func (s *FileService) GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/"
	}
	return homeDir
}

// BrowseDirectory 浏览目录
func (s *FileService) BrowseDirectory(dirPath string) (string, []string, error) {
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
		return dirPath, nil, fmt.Errorf("目录不存在")
	}

	if !info.IsDir() {
		return dirPath, nil, fmt.Errorf("不是目录")
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return dirPath, nil, err
	}

	var directories []string
	// 添加父目录
	if dirPath != "/" {
		directories = append(directories, "..")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return dirPath, directories, nil
}
