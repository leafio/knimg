package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	webview "github.com/webview/webview_go"
	"knimg/handlers"
)

// getBaseDir 获取基础目录，区分开发和生产模式
func getBaseDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		logAndExit(fmt.Sprintf("错误: 无法获取当前目录: %v\n", err))
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("警告: 无法获取可执行文件路径: %v\n", err)
		return currentDir
	}

	execDir := filepath.Dir(execPath)
	isDevMode := strings.Contains(execPath, "go-build") ||
		strings.Contains(execPath, "/Temp/") ||
		strings.Contains(execPath, "\\Temp\\")

	var baseDir string
	if isDevMode {
		baseDir = currentDir
		fmt.Printf("开发模式 - 使用当前工作目录: %s\n", baseDir)
	} else {
		baseDir = execDir
		fmt.Printf("生产模式 - 使用可执行文件目录: %s\n", baseDir)
	}

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		logAndExit(fmt.Sprintf("错误: 可执行文件目录不存在: %s\n", baseDir))
	}

	return baseDir
}

// logAndExit 记录错误并退出程序
func logAndExit(msg string) {
	fmt.Print(msg)
	os.Exit(1)
}

// InitServer 初始化服务器配置
func InitServer(isDevMode bool) (http.Handler, string) {
	baseDir := getBaseDir()

	logPath := filepath.Join(baseDir, "knimg.log")
	if logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		log.SetOutput(logFile)
	} else {
		fmt.Printf("无法创建日志文件: %v\n", err)
	}

	fileHandler := handlers.NewFileHandler(baseDir)
	compressHandler := handlers.NewCompressHandler(baseDir)

	mux := http.NewServeMux()

	if isDevMode {
		frontendDir := filepath.Join(baseDir, "frontend")
		mux.Handle("/", http.FileServer(http.Dir(frontendDir)))
		fmt.Println("✓ 本地前端资源加载成功")
		log.Println("✓ 本地前端资源加载成功")
	} else {
		mux.Handle("/", http.FileServer(GetEmbeddedFrontend()))
		fmt.Println("✓ 嵌入前端资源加载成功")
		log.Println("✓ 嵌入前端资源加载成功")
	}

	mux.HandleFunc("/api/files", fileHandler.ListFiles)
	mux.HandleFunc("/api/files/export", fileHandler.ExportFiles)
	mux.HandleFunc("/api/file/image", fileHandler.ServeImage)
	mux.HandleFunc("/api/directory/home", fileHandler.GetHomeDirectory)
	mux.HandleFunc("/api/directory/browse", fileHandler.BrowseDirectory)
	mux.HandleFunc("/api/compress", compressHandler.CompressFiles)
	mux.HandleFunc("/api/compress/stats", compressHandler.GetCompressionStats)

	return mux, baseDir
}

// StartServer 启动服务器并处理端口冲突
func StartServer(mux http.Handler, baseDir string) (string, webview.WebView) {
	var listener net.Listener
	port := 0
	for p := 8080; p < 8090; p++ {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if err != nil {
			continue
		}
		listener = l
		port = p
		break
	}
	if listener == nil {
		logAndExit("错误: 无法找到可用端口 (8080-8089)\n")
	}

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("服务器启动失败: %v", err)
		}
	}()

	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Printf("🚀 服务器启动在 %s", addr)
	fmt.Printf("🚀 服务器启动在 %s\n", addr)
	log.Printf("📁 工作目录：%s", baseDir)
	fmt.Printf("📁 工作目录：%s\n", baseDir)

	w := webview.New(false)
	w.SetTitle("KnImg")
	w.SetSize(1024, 768, webview.HintNone)
	// 用 localhost 而非 127.0.0.1：系统代理通常会绕过 localhost，但可能拦截字面 IP
	w.Navigate(fmt.Sprintf("http://localhost:%d", port))

	return addr, w
}
