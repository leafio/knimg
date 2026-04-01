# KnImg - 文件管理与图片压缩工具

一个功能强大的本地文件管理和图片压缩工具，基于 Go 语言开发，提供美观的 Web 界面。

## ✨ 功能特性

### 📁 文件管理
- **智能文件夹选择**：友好的文件夹浏览对话框，支持选择工作目录和导出目录
- **递归扫描**：自动扫描指定目录下的所有文件
- **实时统计**：显示各类文件数量（图片、文档、视频、其他）
- **详细信息**：文件名、路径、大小、修改时间、扩展名

### 🔍 智能筛选
- **名称搜索**：模糊匹配文件名
- **文件类型**：按图片/文档/视频/其他筛选
- **自定义扩展名**：支持 `.js`, `.ts`, `.py` 等自定义扩展名筛选，多个扩展名用逗号分隔
- **大小范围**：按文件大小范围筛选
- **排序功能**：支持按名称、大小、类型、修改时间排序

### 📤 多格式导出
- **Excel (.xlsx)**：格式化表格，包含扩展名信息
- **CSV (.csv)**：通用数据格式
- **JSON (.json)**：结构化数据
- **筛选导出**：支持导出时应用当前筛选条件

### 🖼️ 图片压缩
- **格式支持**：JPEG、PNG、GIF、WebP
- **智能压缩策略**：根据文件大小自动调整质量
- **WebP 转换**：WebP 图片自动转换为 JPEG 以获得更好兼容性
- **质量可调**：1-100 可调节压缩质量
- **实时统计**：显示压缩结果和压缩率

## 🚀 快速开始

### 方式一：使用预编译版本

1. 从 [dist/](dist/) 目录下载适合你系统的版本
2. 解压后运行可执行文件
3. 浏览器访问 http://localhost:8080

### 方式二：从源码运行

```bash
# 克隆项目
git clone <repository-url>
cd knimg

# 安装依赖
go mod download

# 运行
go run main.go

# 或编译后运行
go build -o knimg main.go
./knimg
```

### 方式三：开发模式

```bash
# 设置 Go 环境（如需要）
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:$PATH

# 运行
go run main.go
```

访问 http://localhost:8080 开始使用！

## 📖 使用说明

### 1. 选择工作目录
- 点击「工作目录」输入框旁的「📁 浏览」按钮
- 在弹出的对话框中选择要分析的文件夹
- 或直接在输入框中输入文件夹路径

### 2. 文件筛选
- **名称搜索**：在搜索框中输入文件名关键词
- **文件类型**：从下拉菜单选择文件类型
- **自定义扩展名**：输入扩展名，如 `js,ts,py`（支持带点或不带点）
- **大小范围**：设置最小和最大文件大小（KB）
- **排序**：选择排序字段和排序方式

### 3. 图片压缩
- 勾选要压缩的图片文件（只有图片文件可勾选）
- 调整压缩质量滑块（1-100）
- 点击「🗜️ 压缩选中」按钮

### 4. 导出文件
- 选择导出目录（可选）
- 点击对应格式按钮导出：
  - 📊 Excel - 适合人工查看
  - 📄 CSV - 通用性强
  - 📋 JSON - 程序友好

## 📦 项目结构

```
knimg/
├── main.go                    # Web 服务主程序
├── go.mod                     # Go 模块依赖
├── go.sum                     # 依赖锁定文件
├── .gitignore                 # Git 配置
├── readme.md                  # 项目文档
├── models/
│   └── file_models.go         # 数据模型定义
├── handlers/
│   ├── file_handlers.go       # 文件管理 API
│   └── compress_handler.go    # 图片压缩 API
├── frontend/
│   └── index.html             # 前端单页应用
└── dist/                      # 打包文件目录
    ├── knimg-windows.zip      # Windows 版本
    ├── knimg-macos-amd64.zip  # macOS Intel 版本
    └── knimg-macos-arm64.zip  # macOS Apple Silicon 版本
```

## 🛠️ 技术栈

- **后端**：Go + Gin Web 框架
- **前端**：原生 JavaScript + HTML5 + CSS3
- **图片处理**：标准库 `image` 包 + `golang.org/x/image`
- **Excel 导出**：tealeg/xlsx
- **跨平台**：支持 Windows、macOS、Linux

## 📝 开发笔记

### 编译命令

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o knimg.exe main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o knimg main.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o knimg main.go
```

### Git 提交规范

```bash
git add .
git commit -m "feat: 描述新功能"
git commit -m "fix: 描述修复"
git commit -m "docs: 更新文档"
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**构建日期**：2026-04-01  
**Go 版本**：1.26  
**项目状态**：✅ 可用于生产环境
