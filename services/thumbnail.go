package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"strings"
	"sync"

	_ "golang.org/x/image/bmp" // 注册 bmp 解码
	"golang.org/x/image/draw"
)

// ThumbMaxSize 缩略图最长边像素
const ThumbMaxSize = 320

// ServeOriginalAsIs 原文件小于该值(字节)时直接返回原文件，不生成缩略图
const ServeOriginalAsIs = 512 * 1024

// ContentType 返回图片扩展名对应的 MIME 类型，非图片返回空串
func ContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}

// IsSupportedExt 检查扩展名是否为可预览的图片
func IsSupportedExt(ext string) bool {
	return ContentType(ext) != ""
}

// ThumbService 缩略图生成与内存缓存
type ThumbService struct {
	mu      sync.Mutex
	cache   map[string][]byte
	maxSize int
}

// NewThumbService 创建缩略图服务
func NewThumbService() *ThumbService {
	return &ThumbService{
		cache:   make(map[string][]byte),
		maxSize: 512,
	}
}

// Generate 生成缩略图（JPEG 编码，最长边 ThumbMaxSize），按 路径+大小+修改时间 缓存
func (s *ThumbService) Generate(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s|%d|%d", path, info.Size(), info.ModTime().Unix())

	s.mu.Lock()
	if b, ok := s.cache[key]; ok {
		s.mu.Unlock()
		return b, nil
	}
	s.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("无法解码图片: %v", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("无效的图片尺寸")
	}

	scale := float64(ThumbMaxSize) / float64(max(w, h))
	if scale > 1 {
		scale = 1
	}
	dw, dh := max(1, int(float64(w)*scale)), max(1, int(float64(h)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	data := buf.Bytes()

	s.mu.Lock()
	if len(s.cache) >= s.maxSize {
		s.cache = make(map[string][]byte)
	}
	s.cache[key] = data
	s.mu.Unlock()

	return data, nil
}
