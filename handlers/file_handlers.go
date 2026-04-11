package handlers

import (
	"encoding/json"
	"knimg/models"
	"knimg/services"
	"net/http"
	"os"
	"strconv"
)

// FileHandler 文件处理 Web 适配器
type FileHandler struct {
	fileService *services.FileService
}

// NewFileHandler 创建文件处理器
func NewFileHandler(baseDir string) *FileHandler {
	return &FileHandler{
		fileService: services.NewFileService(baseDir),
	}
}

// ListFiles 获取文件列表（支持筛选和排序）
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.FileListRequest
	// 手动绑定查询参数
	req.Search = r.URL.Query().Get("search")
	req.FileType = r.URL.Query().Get("file_type")
	req.FileExt = r.URL.Query().Get("file_ext")
	req.SortBy = r.URL.Query().Get("sort_by")
	req.SortOrder = r.URL.Query().Get("sort_order")

	// 解析大小参数
	if minSize := r.URL.Query().Get("min_size"); minSize != "" {
		if val, err := strconv.ParseInt(minSize, 10, 64); err == nil {
			req.MinSize = val
		}
	}
	if maxSize := r.URL.Query().Get("max_size"); maxSize != "" {
		if val, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			req.MaxSize = val
		}
	}

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
		"success":   true,
		"data":      files,
		"total":     len(files),
		"stats":     stats,
		"work_dir":  workDir,
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

	// 支持自定义导出目录
	exportDir := r.URL.Query().Get("export_dir")

	// 创建导出目录
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无法创建导出目录: " + err.Error(),
		})
		return
	}

	// 获取筛选参数
	var req models.FileListRequest
	// 手动绑定查询参数
	req.Search = r.URL.Query().Get("search")
	req.FileType = r.URL.Query().Get("file_type")
	req.FileExt = r.URL.Query().Get("file_ext")
	req.SortBy = r.URL.Query().Get("sort_by")
	req.SortOrder = r.URL.Query().Get("sort_order")

	// 解析大小参数
	if minSize := r.URL.Query().Get("min_size"); minSize != "" {
		if val, err := strconv.ParseInt(minSize, 10, 64); err == nil {
			req.MinSize = val
		}
	}
	if maxSize := r.URL.Query().Get("max_size"); maxSize != "" {
		if val, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			req.MaxSize = val
		}
	}

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
