//go:build dev

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"knimg/handlers"

	"github.com/gin-gonic/gin"
	webview "github.com/webview/webview_go"
)

func main() {
	// 先输出到控制台
	fmt.Println("=== KnImg 启动 (开发模式) ===")
	fmt.Println("启动时间:", time.Now().Format("2006-01-02 15:04:05"))

	// 尝试创建日志文件
	logPath := "knimg.log"
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Printf("无法创建日志文件: %v\n", err)
		// 继续运行，不使用日志文件
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
		fmt.Printf("日志文件路径: %s\n", logPath)
		log.Printf("日志文件路径: %s", logPath)
	}

	// 记录启动信息
	log.Println("=== KnImg 启动 (开发模式) ===")
	log.Println("启动时间:", time.Now().Format("2006-01-02 15:04:05"))

	var baseDir string

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("错误: 无法获取当前目录: %v\n", err)
		fmt.Println("按任意键退出...")
		var input string
		fmt.Scanln(&input)
		os.Exit(1)
	}

	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("警告: 无法获取可执行文件路径: %v\n", err)
		baseDir = currentDir
		fmt.Printf("使用当前目录: %s\n", baseDir)
		log.Printf("使用当前目录: %s", baseDir)
	} else {
		fmt.Printf("可执行文件路径: %s\n", execPath)
		log.Printf("可执行文件路径: %s", execPath)

		// 开发模式：使用当前工作目录
		baseDir = currentDir
		fmt.Printf("开发模式 - 使用当前工作目录: %s\n", baseDir)
		log.Printf("开发模式 - 使用当前工作目录: %s", baseDir)
	}

	// 确保目录存在
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Printf("错误: 可执行文件目录不存在: %s\n", baseDir)
		fmt.Println("按任意键退出...")
		var input string
		fmt.Scanln(&input)
		os.Exit(1)
	}

	// 创建文件上传目录（如果不存在）
	uploadDir := filepath.Join(baseDir, "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("错误: 无法创建上传目录: %v\n", err)
		fmt.Println("按任意键退出...")
		var input string
		fmt.Scanln(&input)
		os.Exit(1)
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

	// API 路由（先注册，更具体的路径）
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

	// 开发模式：使用本地前端目录
	frontendDir := filepath.Join(baseDir, "frontend")
	log.Printf("开发模式 - 使用本地前端目录: %s", frontendDir)
	fmt.Printf("开发模式 - 使用本地前端目录: %s\n", frontendDir)
	
	// 提供前端静态文件（使用/frontend路径，避免与/api冲突）
	r.Static("/frontend", frontendDir)
	
	// 重定向根路径到/frontend
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/frontend")
	})
	
	fmt.Println("✓ 本地前端资源加载成功")
	log.Println("✓ 本地前端资源加载成功")

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
	w.SetTitle("KnImg (开发模式)")
	w.SetSize(1024, 768, webview.HintNone)
	w.Navigate(fmt.Sprintf("http://localhost:%s/frontend", port))

	// 在goroutine中启动服务器
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 运行WebView主循环
	w.Run()
}
