# KnImg - 文件管理与图片压缩工具

一个本地文件管理和图片压缩工具，基于 Go + WebView 开发，以桌面窗口形式提供美观的 Web 界面。

## ✨ 功能特性

### 📁 文件管理
- **智能文件夹选择**：内置目录浏览器，支持选择工作目录和导出目录
- **递归扫描**：自动扫描指定目录下的所有文件（自动跳过无权限目录）
- **实时统计**：显示各类文件数量（图片、文档、视频、其他）
- **详细信息**：文件名、路径、大小、修改时间、扩展名

### 🔍 智能筛选
- **名称搜索**：模糊匹配文件名（输入即时生效，300ms 防抖）
- **文件类型**：按图片/文档/视频/其他筛选
- **自定义扩展名**：支持 `js,ts,py` 等，多个扩展名用逗号分隔
- **大小范围**：快速预设（>10MB 等）或自定义范围（MB）
- **排序功能**：按名称、大小、类型、修改时间排序

### 📤 多格式导出
- **Excel (.xlsx)** / **CSV (.csv)** / **JSON (.json)**
- **筛选导出**：导出时应用当前筛选条件

### 🖼️ 图片压缩与预览
- **缩略图**：列表内直接显示图片缩略图（大图自动生成 320px 缩略图并缓存，小图原图直出）
- **大图预览**：点击缩略图弹出原尺寸预览，支持 Esc / 点击空白关闭
- **解码支持**：JPEG、PNG、GIF、WebP、BMP（GIF/WebP 压缩时自动转为 JPEG 输出）
- **输出格式**：可选 原格式 / JPEG / PNG（PNG 为无损重编码，不适用质量参数）
- **智能压缩策略**：根据文件大小自动调整质量
- **安全输出**：默认输出到工作目录下的 `compressed/` 文件夹，绝不覆盖原图；覆盖模式采用临时文件 + 原子替换
- **并行压缩**：多核并行处理批量任务（最多 8 并发）
- **质量可调**：10-100 可调节压缩质量

## 🚀 快速开始

### 依赖

- Go 1.25+
- macOS：系统自带 WebKit，无需额外安装
- Windows：WebView2 Runtime（Win11 自带）
- Linux：`libgtk-3-dev libwebkit2gtk-4.0-dev`

### 从源码运行

```bash
git clone <repository-url>
cd knimg
go mod download

# 开发模式：前端从 frontend/ 目录实时读取，改前端代码后刷新窗口即生效
go run -tags dev .

# 生产模式：前端嵌入二进制
go run .
```

### 编译发布

```bash
go build -o knimg .        # 仅生成裸可执行文件（无 .app 包）
./build.sh local           # 只为当前系统构建并打 .app 包，产物在 build/
./build.sh                 # 跨平台打包（Windows/macOS/Linux）
```

## 📖 使用说明

### 1. 选择工作目录
- 点击顶栏「修改」或欢迎页的「选择工作目录」，在内置目录浏览器中选择
- 自动递归扫描所选目录

### 2. 文件筛选
- 名称搜索、扩展名筛选（如 `jpg,png`，带不带点均可）
- 大小快速预设或自定义范围、排序字段与方向切换
- 「清除筛选」一键还原

### 3. 图片压缩
- 勾选要压缩的图片（只有图片可勾选，支持全选）
- 选择输出格式（原格式/JPEG/PNG），留空输出目录则保存到 `compressed/` 子目录
- 调整质量滑块，点击「🗜️ 压缩图片」，完成后显示节省空间比例

### 4. 导出文件
- 「导出」菜单选择格式，在弹出的目录浏览器中确认位置

## 📦 项目结构

```
knimg/
├── main.go                    # 生产模式入口 (//go:build !dev)
├── main_dev.go                # 开发模式入口 (//go:build dev)
├── server.go                  # HTTP 服务与 WebView 窗口启动
├── resources.go               # 前端资源嵌入 (go:embed)
├── server_test.go             # API 集成测试
├── build.sh                   # 跨平台编译脚本
├── models/
│   └── file_models.go         # 数据模型定义
├── handlers/
│   ├── file_handlers.go       # 文件管理 API
│   └── compress_handler.go    # 图片压缩 API
├── services/
│   ├── file_service.go        # 扫描/筛选/导出服务
│   ├── compress_service.go    # 图片压缩服务
│   └── thumbnail.go           # 缩略图生成与缓存
├── frontend/
│   ├── index.html             # 前端单页应用
│   ├── css/                   # 样式（base/layout/components/responsive）
│   └── js/                    # ES Modules（api/app/filters/file-list/compress/utils）
└── build/                     # 打包产物目录
```

## 🛠️ 技术栈

- **后端**：Go 标准库 `net/http` + WebView ([webview_go](https://github.com/webview/webview_go))
- **前端**：原生 JavaScript (ES Modules) + HTML5 + CSS3
- **图片处理**：标准库 `image` 包 + `golang.org/x/image`
- **Excel 导出**：tealeg/xlsx
- **跨平台**：Windows、macOS、Linux

## 📝 开发笔记

### 测试

```bash
go test .        # API 集成测试（扫描/筛选/压缩不覆盖原图/原子写入/导出）
go vet ./...
```

### 运行行为说明

- 服务器仅监听 `127.0.0.1`（不暴露局域网），端口从 8080 起自动探测至 8089
- 日志写入程序所在目录的 `knimg.log`
- 生产模式下前端资源嵌入二进制；开发模式（`-tags dev`）读取磁盘上的 `frontend/`

### Git 提交规范

```bash
git commit -m "feat: 描述新功能"
git commit -m "fix: 描述修复"
git commit -m "docs: 更新文档"
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**更新日期**：2026-08-23  
**Go 版本要求**：1.25+  
**项目状态**：✅ 可用于生产环境
