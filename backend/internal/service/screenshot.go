package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	v1 "backend/api/screenshot/v1"
	"backend/utility/adb"

	"github.com/chai2010/webp"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/image/draw"
)

// 缓存项结构
type screenshotCacheItem struct {
	imageData   string    // Base64编码的图片数据
	timestamp   time.Time // 缓存时间
	quality     int       // 图片质量
	scale       float64   // 图片缩放比例
	format      string    // 图片格式
	contentHash string    // 内容hash，用于判断图片是否变化
}

type IScreenshot interface {
	Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error)
}

var (
	localScreenshot IScreenshot
	screenshotOnce  sync.Once
)

// Screenshot 获取截图服务单例
func Screenshot() IScreenshot {
	screenshotOnce.Do(func() {
		localScreenshot = newScreenshotService()
	})
	return localScreenshot
}

type screenshotService struct {
	cache    map[string]*screenshotCacheItem // 设备ID -> 缓存项
	cacheMux sync.RWMutex                    // 缓存读写锁
	ttl      time.Duration                   // 缓存过期时间

	// 请求合并相关
	inProgress    map[string]chan struct{} // 正在处理中的请求
	inProgressMux sync.RWMutex             // 进行中请求的锁
}

func newScreenshotService() *screenshotService {
	// 从配置文件获取缓存过期时间，默认为5秒
	cacheTTL := g.Cfg().MustGet(context.Background(), "screenshot.cacheTTL", 5).Int()

	s := &screenshotService{
		cache:      make(map[string]*screenshotCacheItem),
		inProgress: make(map[string]chan struct{}),
		ttl:        time.Duration(cacheTTL) * time.Second,
	}
	// 启动定期清理过期缓存的goroutine
	go s.cleanExpiredCache()
	return s
}

// 清理过期缓存
func (s *screenshotService) cleanExpiredCache() {
	ticker := time.NewTicker(time.Second * 30) // 每30秒检查一次
	for range ticker.C {
		s.cacheMux.Lock()
		now := time.Now()
		for deviceID, item := range s.cache {
			if now.Sub(item.timestamp) > s.ttl {
				delete(s.cache, deviceID)
			}
		}
		s.cacheMux.Unlock()
	}
}

// 获取缓存的截图
func (s *screenshotService) getCachedScreenshot(ctx context.Context, deviceID string, quality int, scale float64, format string) *screenshotCacheItem {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	if item, exists := s.cache[deviceID]; exists {
		timeSinceCache := time.Since(item.timestamp)
		if timeSinceCache <= s.ttl &&
			item.quality == quality &&
			item.scale == scale &&
			item.format == format {
			// g.Log().Info(ctx, fmt.Sprintf("设备[%s]命中缓存, 缓存时间: %.2f秒", deviceID, timeSinceCache.Seconds()))
			return item
		}
		if timeSinceCache > s.ttl {
			// g.Log().Info(ctx, fmt.Sprintf("设备[%s]缓存已过期, 缓存时间: %.2f秒", deviceID, timeSinceCache.Seconds()))
		}
		if item.quality != quality {
			// g.Log().Info(ctx, fmt.Sprintf("设备[%s]缓存质量不匹配, 缓存质量: %d, 请求质量: %d", deviceID, item.quality, quality))
		}
	} else {
		// g.Log().Info(ctx, fmt.Sprintf("设备[%s]无缓存", deviceID))
	}
	return nil
}

// 设置缓存
func (s *screenshotService) setCacheScreenshot(ctx context.Context, deviceID string, imageData string, quality int, scale float64, format string, contentHash string) {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	s.cache[deviceID] = &screenshotCacheItem{
		imageData:   imageData,
		timestamp:   time.Now(),
		quality:     quality,
		scale:       scale,
		format:      format,
		contentHash: contentHash,
	}
	// g.Log().Info(ctx, fmt.Sprintf("设备[%s]设置新缓存, 质量: %d", deviceID, quality))
}

// 获取或创建进行中的请求标记
func (s *screenshotService) getOrCreateInProgress(ctx context.Context, deviceID string) (chan struct{}, bool) {
	s.inProgressMux.Lock()
	defer s.inProgressMux.Unlock()

	if ch, exists := s.inProgress[deviceID]; exists {
		// g.Log().Info(ctx, fmt.Sprintf("设备[%s]已有请求在处理中, 等待复用", deviceID))
		return ch, false
	}

	ch := make(chan struct{})
	s.inProgress[deviceID] = ch
	// g.Log().Info(ctx, fmt.Sprintf("设备[%s]开始新的截图请求", deviceID))
	return ch, true
}

// 移除进行中的请求标记
func (s *screenshotService) removeInProgress(deviceID string) {
	s.inProgressMux.Lock()
	defer s.inProgressMux.Unlock()

	if ch, exists := s.inProgress[deviceID]; exists {
		close(ch)
		delete(s.inProgress, deviceID)
	}
}

// 图片缩放函数
func resizeImage(img image.Image, scale float64) image.Image {
	if scale >= 1.0 {
		return img
	}
	bounds := img.Bounds()
	w := int(float64(bounds.Dx()) * scale)
	h := int(float64(bounds.Dy()) * scale)
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	return dst
}

// 计算图片内容hash
func calculateImageHash(img image.Image) string {
	bounds := img.Bounds()
	hasher := md5.New()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			hasher.Write([]byte{byte(r), byte(g), byte(b), byte(a)})
		}
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// 判断图片是否为横屏
// 如果宽度大于高度，则认为是横屏
func isLandscape(img image.Image) bool {
	bounds := img.Bounds()
	return bounds.Dx() > bounds.Dy()
}

// 顺时针旋转图片90度
func rotateImage90CW(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(h-y-1, x, src.At(x, y))
		}
	}
	return dst
}

// 处理图片（包含缩放和旋转）
func processImage(img image.Image, scale float64, autoRotate bool) image.Image {
	// 先进行缩放
	if scale < 1.0 {
		img = resizeImage(img, scale)
	}

	// 如果需要自动旋转且是横屏
	if autoRotate && isLandscape(img) {
		img = rotateImage90CW(img)
	}

	return img
}

// 编码图片为指定格式
func encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	switch format {
	case "webp":
		err = webp.Encode(&buf, img, &webp.Options{
			Lossless: false,
			Quality:  float32(quality),
		})
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{
			Quality: quality,
		})
	default:
		return nil, fmt.Errorf("不支持的图片格式: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *screenshotService) Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error) {
	deviceId := req.DeviceId

	// 从配置文件获取默认值
	defaultQuality := g.Cfg().MustGet(ctx, "screenshot.quality", 80).Int()
	defaultScale := g.Cfg().MustGet(ctx, "screenshot.scale", 1.0).Float64()
	defaultFormat := g.Cfg().MustGet(ctx, "screenshot.format", "webp").String()

	// 使用请求参数或默认值
	quality := req.Quality
	if quality == 0 {
		quality = defaultQuality
	}

	scale := req.Scale
	if scale == 0 {
		scale = defaultScale
	}

	format := req.Format
	if format == "" {
		format = defaultFormat
	}

	// startTime := time.Now()
	// defer func() {
	// 	g.Log().Debug(ctx, fmt.Sprintf("设备[%s]截图请求完成, 耗时: %.2f毫秒, 成功: %v",
	// 		deviceId, float64(time.Since(startTime).Milliseconds()), res.Success))
	// }()

	res = &v1.ScreenshotRes{
		DeviceId:  deviceId,
		Success:   false,
		ImageData: "",
		Error:     "",
	}

	// 尝试从缓存获取
	if cached := s.getCachedScreenshot(ctx, deviceId, quality, scale, format); cached != nil {
		res.Success = true
		res.ImageData = cached.imageData
		return res, nil
	}

	// 检查是否有正在进行的请求
	ch, isFirst := s.getOrCreateInProgress(ctx, deviceId)
	if !isFirst {
		select {
		case <-ch:
			if cached := s.getCachedScreenshot(ctx, deviceId, quality, scale, format); cached != nil {
				res.Success = true
				res.ImageData = cached.imageData
				return res, nil
			}
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}

	defer s.removeInProgress(deviceId)

	adbTool := adb.Default()

	if err = adbTool.Connect(deviceId); err != nil {
		res.Error = fmt.Sprintf("连接设备失败: %v", err)
		return res, nil
	}

	timestamp := time.Now().UnixNano()
	screenshotDir := "uploads/screenshots"
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		res.Error = fmt.Sprintf("创建截图目录失败: %v", err)
		return res, nil
	}

	tempFileName := filepath.Join(screenshotDir, fmt.Sprintf("screenshot_%s_%d.png", strings.Replace(deviceId, ":", "_", -1), timestamp))
	deviceTempPath := fmt.Sprintf("/data/local/tmp/%s", filepath.Base(tempFileName))

	if err = adbTool.Screencap(deviceId, deviceTempPath); err != nil {
		res.Error = fmt.Sprintf("设备截图失败: %v", err)
		return res, nil
	}

	if err = adbTool.PullFile(deviceId, deviceTempPath, tempFileName); err != nil {
		res.Error = fmt.Sprintf("拉取截图失败: %v", err)
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		return res, nil
	}

	imgData, err := os.ReadFile(tempFileName)
	if err != nil {
		res.Error = fmt.Sprintf("读取截图文件失败: %v", err)
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return res, nil
	}

	img, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		res.Error = fmt.Sprintf("解码PNG图片失败: %v", err)
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return res, nil
	}

	contentHash := calculateImageHash(img)

	if cached := s.getCachedScreenshot(ctx, deviceId, quality, scale, format); cached != nil {
		if cached.contentHash == contentHash {
			res.Success = true
			res.ImageData = cached.imageData
			return res, nil
		}
	}

	// 处理图片（缩放和旋转）
	img = processImage(img, scale, true) // true表示启用自动旋转

	// 编码图片
	encodedData, err := encodeImage(img, format, quality)
	if err != nil {
		res.Error = fmt.Sprintf("图片编码失败: %v", err)
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return res, nil
	}

	// Base64编码
	imageData := fmt.Sprintf("data:image/%s;base64,%s",
		format,
		base64.StdEncoding.EncodeToString(encodedData))

	// 更新缓存
	s.setCacheScreenshot(ctx, deviceId, imageData, quality, scale, format, contentHash)

	res.Success = true
	res.ImageData = imageData

	// 清理临时文件
	adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
	os.Remove(tempFileName)

	return res, nil
}
