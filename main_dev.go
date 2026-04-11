//go:build dev

package main

import (
	"fmt"
	"time"
)

func main() {
	// 初始化日志
	fmt.Println("=== KnImg 启动 (开发模式) ===")
	fmt.Println("启动时间:", time.Now().Format("2006-01-02 15:04:05"))

	// 初始化服务器
	mux, baseDir := InitServer(true)

	// 启动服务器并创建WebView窗口
	_, w := StartServer(mux, baseDir)
	defer w.Destroy()
	w.SetTitle("KnImg (开发模式)")

	// 运行WebView主循环
	w.Run()
}
