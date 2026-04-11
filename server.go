package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"knimg/handlers"
	webview "github.com/webview/webview_go"
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
		fmt.Printf("使用当前目录: %s\n", currentDir)
		log.Printf("使用当前目录: %s", currentDir)
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
		log.Printf("开发模式 - 使用当前工作目录: %s", baseDir)
	} else {
		baseDir = execDir
		fmt.Printf("生产模式 - 使用可执行文件目录: %s\n", baseDir)
		log.Printf("生产模式 - 使用可执行文件目录: %s", baseDir)
	}

	// 确保目录存在
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		logAndExit(fmt.Sprintf("错误: 可执行文件目录不存在: %s\n", baseDir))
	}

	return baseDir
}

// logAndExit 记录错误并退出程序
func logAndExit(msg string) {
	fmt.Print(msg)
	fmt.Println("按任意键退出...")
	var input string
	fmt.Scanln(&input)
	os.Exit(1)
}

// withCORS 添加 CORS 中间件
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next(w, r)
	}
}

// InitServer 初始化服务器配置
func InitServer(isDevMode bool) (http.Handler, string) {
	// 初始化日志
	logPath := "knimg.log"
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Printf("无法创建日志文件: %v\n", err)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
		fmt.Printf("日志文件路径: %s\n", logPath)
	}

	// 获取基础目录
	baseDir := getBaseDir()

	// 初始化处理器
	fileHandler := handlers.NewFileHandler(baseDir)
	compressHandler := handlers.NewCompressHandler(baseDir)

	// 创建多路复用器
	mux := http.NewServeMux()



	// 处理前端资源
	if isDevMode {
		// 开发模式：使用本地前端目录
		frontendDir := filepath.Join(baseDir, "frontend")
		log.Printf("开发模式 - 使用本地前端目录: %s", frontendDir)
		fmt.Printf("开发模式 - 使用本地前端目录: %s\n", frontendDir)

		// 提供前端静态文件
		frontendHandler := http.FileServer(http.Dir(frontendDir))
		mux.Handle("/", frontendHandler)
		fmt.Println("✓ 本地前端资源加载成功")
		log.Println("✓ 本地前端资源加载成功")
	} else {
		// 生产模式：使用嵌入的前端资源
		fmt.Println("✓ 嵌入前端资源加载成功")
		log.Println("✓ 嵌入前端资源加载成功")

		// 提供嵌入的前端index.html
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write(indexHTMLContent)
			} else {
				http.NotFound(w, r)
			}
		})
	}

	// API 路由
	mux.HandleFunc("/api/files", withCORS(fileHandler.ListFiles))
	mux.HandleFunc("/api/files/export", withCORS(fileHandler.ExportFiles))
	mux.HandleFunc("/api/directory/home", withCORS(fileHandler.GetHomeDirectory))
	mux.HandleFunc("/api/directory/browse", withCORS(fileHandler.BrowseDirectory))
	mux.HandleFunc("/api/compress", withCORS(compressHandler.CompressFiles))
	mux.HandleFunc("/api/compress/stats", withCORS(compressHandler.GetCompressionStats))

	return mux, baseDir
}

// StartServer 启动服务器并处理端口冲突
func StartServer(mux http.Handler, baseDir string) (string, webview.WebView) {
	port := "8080"
	serverAddr := ":" + port
	var server *http.Server

	// 尝试启动服务器，如果端口被占用，尝试其他端口
	for i := 0; i < 10; i++ {
		// 先尝试绑定端口，检查是否被占用
		listener, err := net.Listen("tcp", serverAddr)
		if err != nil {
			// 端口被占用，尝试下一个端口
			log.Printf("端口 %s 被占用，尝试下一个端口", port)
			fmt.Printf("端口 %s 被占用，尝试下一个端口\n", port)
			port = fmt.Sprintf("%d", 8080+i+1)
			serverAddr = ":" + port
			continue
		}
		
		// 端口可用，关闭监听器
		listener.Close()
		
		// 创建新的服务器实例
		server = &http.Server{
			Addr:    serverAddr,
			Handler: mux,
		}
		
		// 在goroutine中启动服务器
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("服务器启动失败: %v", err)
			}
		}()
		
		// 等待服务器启动
		time.Sleep(300 * time.Millisecond)
		
		// 检查服务器是否成功启动
		conn, err := net.Dial("tcp", "localhost"+serverAddr)
		if err == nil {
			conn.Close()
			// 服务器启动成功
			log.Printf("🚀 服务器启动在 http://localhost%s", serverAddr)
			fmt.Printf("🚀 服务器启动在 http://localhost%s\n", serverAddr)
			log.Printf("📁 工作目录：%s", baseDir)
			fmt.Printf("📁 工作目录：%s\n", baseDir)
			break
		}
		
		// 服务器启动失败，尝试下一个端口
		if server != nil {
			// 创建一个上下文来关闭服务器
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		}
		
		log.Printf("服务器启动失败，尝试下一个端口")
		fmt.Printf("服务器启动失败，尝试下一个端口\n")
		port = fmt.Sprintf("%d", 8080+i+1)
		serverAddr = ":" + port
	}

	// 创建WebView窗口
	w := webview.New(false)
	w.SetTitle("KnImg")
	w.SetSize(1024, 768, webview.HintNone)
	w.Navigate(fmt.Sprintf("http://localhost%s", serverAddr))

	return serverAddr, w
}
