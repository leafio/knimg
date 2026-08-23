package models

// FileInfo 文件信息结构体
type FileInfo struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Type       string `json:"type"`
	Ext        string `json:"ext"` // 文件扩展名
	ModTime    string `json:"mod_time"`
	Compressed bool   `json:"compressed"`
}

// FileListRequest 文件列表请求参数
type FileListRequest struct {
	Search    string `json:"search" form:"search"`         // 名称模糊搜索
	FileType  string `json:"file_type" form:"file_type"`   // 文件类型筛选
	FileExt   string `json:"file_ext" form:"file_ext"`     // 自定义扩展名筛选（多个用逗号分隔）
	MinSize   int64  `json:"min_size" form:"min_size"`     // 最小大小（字节）
	MaxSize   int64  `json:"max_size" form:"max_size"`     // 最大大小（字节）
	SortBy    string `json:"sort_by" form:"sort_by"`       // 排序字段
	SortOrder string `json:"sort_order" form:"sort_order"` // 排序方式 asc/desc
}

// CompressRequest 压缩请求参数
type CompressRequest struct {
	Files     []string `json:"files"`
	Quality   int      `json:"quality"`
	Format    string   `json:"format"` // 目标格式：空=原格式，jpeg/png
	OutputDir string   `json:"output_dir"`
	WorkDir   string   `json:"work_dir"`  // 自定义工作目录
	Overwrite bool     `json:"overwrite"` // 是否覆盖原文件
}
