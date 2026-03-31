package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"knimg/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取工作目录:", err)
	}

	// 设置基础目录为当前目录
	baseDir := workDir

	// 创建文件上传目录（如果不存在）
	uploadDir := filepath.Join(baseDir, "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("无法创建上传目录:", err)
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
	r.Static("/compressed", filepath.Join(baseDir, "compressed"))
	r.StaticFile("/", "./frontend/index.html")

	// 启动服务器
	port := "8080"
	fmt.Printf("🚀 服务器启动在 http://localhost:%s\n", port)
	fmt.Printf("📁 工作目录：%s\n", baseDir)
	fmt.Printf("📤 上传目录：%s\n", uploadDir)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
