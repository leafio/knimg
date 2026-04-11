//go:build !dev

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"knimg/handlers"

	"github.com/gin-gonic/gin"
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

func main() {
	// 初始化日志
	fmt.Println("=== KnImg 启动 ===")
	fmt.Println("启动时间:", time.Now().Format("2006-01-02 15:04:05"))

	logPath := "knimg.log"
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Printf("无法创建日志文件: %v\n", err)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
		fmt.Printf("日志文件路径: %s\n", logPath)
	}
	log.Println("=== KnImg 启动 ===")
	log.Println("启动时间:", time.Now().Format("2006-01-02 15:04:05"))

	// 获取基础目录
	baseDir := getBaseDir()

	// 创建必要的目录
	uploadDir := filepath.Join(baseDir, "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logAndExit(fmt.Sprintf("错误: 无法创建上传目录: %v\n", err))
	}

	// 初始化处理器
	fileHandler := handlers.NewFileHandler(baseDir)
	compressHandler := handlers.NewCompressHandler(baseDir)

	// 创建 Gin 路由器
	r := gin.Default()

	// 启用 CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
	})

	// API 路由
	api := r.Group("/api")
	{
		// 文件列表
		api.GET("/files", fileHandler.ListFiles)

		// 导出文件列表
		api.GET("/files/export", fileHandler.ExportFiles)

		// 目录浏览
		api.GET("/directory/home", fileHandler.GetHomeDirectory)
		api.GET("/directory/browse", fileHandler.BrowseDirectory)

		// 批量压缩图片
		api.POST("/compress", compressHandler.CompressFiles)

		// 获取压缩统计
		api.GET("/compress/stats", compressHandler.GetCompressionStats)
	}

	// 静态文件服务
	r.Static("/uploads", uploadDir)
	compressedDir := filepath.Join(baseDir, "compressed")
	log.Printf("压缩目录: %s", compressedDir)
	if err := os.MkdirAll(compressedDir, 0755); err != nil {
		log.Fatalf("无法创建压缩目录: %v", err)
	}
	r.Static("/compressed", compressedDir)

	// 检查是否为开发模式
	execPath, _ := os.Executable()
	isDevMode := strings.Contains(execPath, "go-build") ||
		strings.Contains(execPath, "/Temp/") ||
		strings.Contains(execPath, "\\Temp\\")

	if isDevMode {
		// 开发模式：使用本地前端目录
		frontendDir := filepath.Join(baseDir, "frontend")
		log.Printf("开发模式 - 使用本地前端目录: %s", frontendDir)
		fmt.Printf("开发模式 - 使用本地前端目录: %s\n", frontendDir)

		// 提供前端静态文件
		r.Static("/", frontendDir)
		fmt.Println("✓ 本地前端资源加载成功")
		log.Println("✓ 本地前端资源加载成功")
	} else {
		// 生产模式：使用嵌入的前端资源
		fmt.Println("✓ 嵌入前端资源加载成功")
		log.Println("✓ 嵌入前端资源加载成功")

		// 提供嵌入的前端index.html
		r.GET("/", func(c *gin.Context) {
			// 生产模式下会使用嵌入的资源
			c.Data(200, "text/html; charset=utf-8", indexHTMLContent)
		})
	}

	// 启动服务器
	port := "8080"
	log.Printf("🚀 服务器启动在 http://localhost:%s", port)
	log.Printf("📁 工作目录：%s", baseDir)
	log.Printf("📤 上传目录：%s", uploadDir)

	// 同时输出到控制台
	fmt.Printf("🚀 服务器启动在 http://localhost:%s\n", port)
	fmt.Printf("📁 工作目录：%s\n", baseDir)
	fmt.Printf("📤 上传目录：%s\n", uploadDir)

	// 创建WebView窗口
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("KnImg")
	w.SetSize(1024, 768, webview.HintNone)
	w.Navigate(fmt.Sprintf("http://localhost:%s", port))

	// 在goroutine中启动服务器
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 运行WebView主循环
	w.Run()
}
