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
)

func main() {
	// 先输出到控制台
	fmt.Println("=== KnImg 启动 ===")
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
	log.Println("=== KnImg 启动 ===")
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

		execDir := filepath.Dir(execPath)
		
		// 检查是否在临时目录中运行（开发模式）
		isDevMode := strings.Contains(execPath, "go-build") || 
		             strings.Contains(execPath, "/Temp/") ||
		             strings.Contains(execPath, "\\Temp\\")
		
		if isDevMode {
			// 开发模式：使用当前工作目录
			baseDir = currentDir
			fmt.Printf("开发模式 - 使用当前工作目录: %s\n", baseDir)
			log.Printf("开发模式 - 使用当前工作目录: %s", baseDir)
		} else {
			// 生产模式：使用可执行文件目录
			baseDir = execDir
			fmt.Printf("生产模式 - 使用可执行文件目录: %s\n", baseDir)
			log.Printf("生产模式 - 使用可执行文件目录: %s", baseDir)
		}
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

	// 检查前端目录
	frontendDir := filepath.Join(baseDir, "frontend")
	fmt.Printf("前端目录: %s\n", frontendDir)
	log.Printf("前端目录: %s", frontendDir)
	
	// 如果前端目录不存在，尝试使用当前工作目录
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		fmt.Printf("警告: 前端目录不存在: %s\n", frontendDir)
		fmt.Println("尝试使用当前工作目录...")
		
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("错误: 无法获取当前目录: %v\n", err)
			fmt.Println("按任意键退出...")
			var input string
			fmt.Scanln(&input)
			os.Exit(1)
		}
		
		frontendDir = filepath.Join(currentDir, "frontend")
		fmt.Printf("尝试前端目录: %s\n", frontendDir)
		log.Printf("尝试前端目录: %s", frontendDir)
		
		if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
			fmt.Printf("错误: 前端目录不存在: %s\n", frontendDir)
			fmt.Println("请确保 frontend 文件夹与 knimg.exe 在同一目录")
			fmt.Println("按任意键退出...")
			var input string
			fmt.Scanln(&input)
			os.Exit(1)
		}
	}

	frontendIndex := filepath.Join(frontendDir, "index.html")
	fmt.Printf("前端index.html路径: %s\n", frontendIndex)
	log.Printf("前端index.html路径: %s", frontendIndex)
	if _, err := os.Stat(frontendIndex); os.IsNotExist(err) {
		fmt.Printf("错误: 前端index.html文件不存在: %s\n", frontendIndex)
		fmt.Println("请确保 frontend/index.html 文件存在")
		fmt.Println("按任意键退出...")
		var input string
		fmt.Scanln(&input)
		os.Exit(1)
	}

	// 静态文件服务
	r.Static("/uploads", uploadDir)
	compressedDir := filepath.Join(baseDir, "compressed")
	log.Printf("压缩目录: %s", compressedDir)
	if err := os.MkdirAll(compressedDir, 0755); err != nil {
		log.Fatalf("无法创建压缩目录: %v", err)
	}
	r.Static("/compressed", compressedDir)
	r.StaticFile("/", frontendIndex)

	// 启动服务器
	port := "8080"
	log.Printf("🚀 服务器启动在 http://localhost:%s", port)
	log.Printf("📁 工作目录：%s", baseDir)
	log.Printf("📤 上传目录：%s", uploadDir)

	// 同时输出到控制台
	fmt.Printf("🚀 服务器启动在 http://localhost:%s\n", port)
	fmt.Printf("📁 工作目录：%s\n", baseDir)
	fmt.Printf("📤 上传目录：%s\n", uploadDir)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
