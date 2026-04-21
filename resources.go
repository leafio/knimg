package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed frontend/*
var embeddedFS embed.FS

// GetEmbeddedFrontend 获取嵌入的前端文件系统
func GetEmbeddedFrontend() http.FileSystem {
	subFS, err := fs.Sub(embeddedFS, "frontend")
	if err != nil {
		panic(err)
	}
	return http.FS(subFS)
}
