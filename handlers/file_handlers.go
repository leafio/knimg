package handlers

import (
	"encoding/json"
	"fmt"
	"knimg/models"
	"knimg/services"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileHandler 文件处理 Web 适配器
type FileHandler struct {
	fileService *services.FileService
	thumbSvc    *services.ThumbService
}

// NewFileHandler 创建文件处理器
func NewFileHandler(baseDir string) *FileHandler {
	return &FileHandler{
		fileService: services.NewFileService(baseDir),
		thumbSvc:    services.NewThumbService(),
	}
}

// resolveUnderWorkDir 将相对路径解析为 workDir 下的绝对路径，防止路径越界
func resolveUnderWorkDir(workDir, relPath string) (string, error) {
	if workDir == "" || relPath == "" {
		return "", fmt.Errorf("参数不完整")
	}
	cleanWork := filepath.Clean(workDir)
	full := filepath.Clean(filepath.Join(cleanWork, relPath))
	if full != cleanWork && !strings.HasPrefix(full, cleanWork+string(os.PathSeparator)) {
		return "", fmt.Errorf("非法路径")
	}
	return full, nil
}

// ServeImage 提供图片内容：小图/原图请求直接返回原文件，大图返回缩略图
func (h *FileHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	workDir := q.Get("work_dir")
	full := q.Get("full") == "1"

	fullPath, err := resolveUnderWorkDir(workDir, q.Get("path"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ext := filepath.Ext(fullPath)
	contentType := services.ContentType(ext)
	if contentType == "" {
		http.Error(w, "不支持的图片格式", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	etag := fmt.Sprintf(`"%d-%d"`, info.Size(), info.ModTime().Unix())
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	if match := r.Header.Get("If-None-Match"); match == etag && !full {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if full || info.Size() <= services.ServeOriginalAsIs {
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, fullPath)
		return
	}

	thumb, err := h.thumbSvc.Generate(fullPath)
	if err != nil {
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, fullPath)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(thumb)
}

// parseListRequest 从查询参数解析筛选请求
func parseListRequest(r *http.Request) models.FileListRequest {
	q := r.URL.Query()
	req := models.FileListRequest{
		Search:    q.Get("search"),
		FileType:  q.Get("file_type"),
		FileExt:   q.Get("file_ext"),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
	}
	if val, err := strconv.ParseInt(q.Get("min_size"), 10, 64); err == nil {
		req.MinSize = val
	}
	if val, err := strconv.ParseInt(q.Get("max_size"), 10, 64); err == nil {
		req.MaxSize = val
	}
	return req
}

// ListFiles 获取文件列表（支持筛选和排序）
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := parseListRequest(r)
	workDir := h.fileService.GetWorkDir(r.URL.Query().Get("work_dir"))

	files, err := h.fileService.ScanFilesWithFilter(workDir, &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 计算统计信息
	stats := h.fileService.CalculateStats(files)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"data":     files,
		"total":    len(files),
		"stats":    stats,
		"work_dir": workDir,
	})
}

// ExportFiles 导出文件列表
func (h *FileHandler) ExportFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "excel"
	}

	exportDir := r.URL.Query().Get("export_dir")
	if exportDir != "" {
		if err := os.MkdirAll(exportDir, 0755); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "无法创建导出目录: " + err.Error(),
			})
			return
		}
	}

	req := parseListRequest(r)
	workDir := h.fileService.GetWorkDir(r.URL.Query().Get("work_dir"))

	files, err := h.fileService.ScanFilesWithFilter(workDir, &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var filePath string
	switch format {
	case "excel":
		filePath, err = h.fileService.ExportToExcel(files, exportDir)
	case "csv":
		filePath, err = h.fileService.ExportToCSV(files, exportDir)
	case "json":
		filePath, err = h.fileService.ExportToJSON(files, exportDir)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "不支持的导出格式",
		})
		return
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"file_path":  filePath,
		"export_dir": exportDir,
	})
}

// GetHomeDirectory 获取用户主目录
func (h *FileHandler) GetHomeDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	homeDir := h.fileService.GetHomeDirectory()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"path":    homeDir,
	})
}

// BrowseDirectory 浏览目录
func (h *FileHandler) BrowseDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dirPath := r.URL.Query().Get("path")
	currentPath, directories, err := h.fileService.BrowseDirectory(dirPath)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"current_path": currentPath,
		"directories":  directories,
	})
}
