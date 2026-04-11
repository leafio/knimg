package handlers

import (
	"encoding/json"
	"knimg/models"
	"knimg/services"
	"net/http"
)

// CompressHandler 图片压缩处理 Web 适配器
type CompressHandler struct {
	compressService *services.CompressService
}

// NewCompressHandler 创建压缩处理器
func NewCompressHandler(baseDir string) *CompressHandler {
	return &CompressHandler{
		compressService: services.NewCompressService(baseDir),
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

	// 调用压缩服务
	result := h.compressService.CompressFiles(req.Files, req.Quality, req.WorkDir, req.OutputDir, req.Overwrite)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// GetCompressionStats 获取压缩统计信息
func (h *CompressHandler) GetCompressionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 支持自定义压缩目录
	compressedDir := r.URL.Query().Get("compressed_dir")

	// 调用压缩服务获取统计信息
	totalFiles, totalSize := h.compressService.GetCompressionStats(compressedDir)

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
