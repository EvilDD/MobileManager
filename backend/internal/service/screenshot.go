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
	s := &screenshotService{
		cache:      make(map[string]*screenshotCacheItem),
		inProgress: make(map[string]chan struct{}),
		ttl:        time.Second * 5, // 增加到5秒的缓存过期时间
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

func (s *screenshotService) Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error) {
	deviceId := req.DeviceId
	quality := req.Quality
	scale := req.Scale
	format := req.Format

	if quality == 0 {
		quality = 80 // 默认质量
	}
	if scale == 0 {
		scale = 1.0 // 默认不缩放
	}
	if format == "" {
		format = "webp" // 默认使用WebP格式
	}

	startTime := time.Now()
	defer func() {
		g.Log().Debug(ctx, fmt.Sprintf("设备[%s]截图请求完成, 耗时: %.2f毫秒, 成功: %v",
			deviceId, float64(time.Since(startTime).Milliseconds()), res.Success))
	}()

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
		// 等待已有请求完成
		select {
		case <-ch:
			// 其他请求已完成，尝试从缓存获取
			if cached := s.getCachedScreenshot(ctx, deviceId, quality, scale, format); cached != nil {
				res.Success = true
				res.ImageData = cached.imageData
				return res, nil
			}
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}

	// 确保在函数返回时移除进行中标记
	defer s.removeInProgress(deviceId)

	// 获取 ADB 工具实例
	adbTool := adb.Default()

	// 1. 连接设备
	if err = adbTool.Connect(deviceId); err != nil {
		res.Error = fmt.Sprintf("连接设备失败: %v", err)
		return res, nil
	}

	// 生成唯一的临时文件名
	timestamp := time.Now().UnixNano()

	// 确保资源目录存在
	screenshotDir := "resource/screenshots"
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		res.Error = fmt.Sprintf("创建截图目录失败: %v", err)
		return res, nil
	}

	tempFileName := filepath.Join(screenshotDir, fmt.Sprintf("screenshot_%s_%d.png", strings.Replace(deviceId, ":", "_", -1), timestamp))
	deviceTempPath := fmt.Sprintf("/data/local/tmp/%s", filepath.Base(tempFileName))

	// 2. 在设备上执行截图命令
	if err = adbTool.Screencap(deviceId, deviceTempPath); err != nil {
		res.Error = fmt.Sprintf("设备截图失败: %v", err)
		return res, nil
	}

	// 3. 从设备中拉取截图文件
	if err = adbTool.PullFile(deviceId, deviceTempPath, tempFileName); err != nil {
		res.Error = fmt.Sprintf("拉取截图失败: %v", err)
		// 清理设备上的临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		return res, nil
	}

	// 4. 读取本地文件
	imgData, err := os.ReadFile(tempFileName)
	if err != nil {
		res.Error = fmt.Sprintf("读取截图文件失败: %v", err)
		// 清理设备上和本地的临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return res, nil
	}

	// 5. 解码PNG图片
	img, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		res.Error = fmt.Sprintf("解码PNG图片失败: %v", err)
		// 清理临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return res, nil
	}

	// 计算原始图片hash
	contentHash := calculateImageHash(img)

	// 检查是否需要更新缓存
	if cached := s.getCachedScreenshot(ctx, deviceId, quality, scale, format); cached != nil {
		if cached.contentHash == contentHash {
			res.Success = true
			res.ImageData = cached.imageData
			return res, nil
		}
	}

	// 6. 缩放图片
	if scale < 1.0 {
		img = resizeImage(img, scale)
	}

	// 7. 根据格式进行编码
	var buf bytes.Buffer
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
		err = fmt.Errorf("不支持的图片格式: %s", format)
	}

	if err != nil {
		res.Error = fmt.Sprintf("图片编码失败: %v", err)
		// 清理临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return res, nil
	}

	// 8. Base64编码
	imageData := fmt.Sprintf("data:image/%s;base64,%s",
		format,
		base64.StdEncoding.EncodeToString(buf.Bytes()))

	// 9. 更新缓存
	s.setCacheScreenshot(ctx, deviceId, imageData, quality, scale, format, contentHash)

	res.Success = true
	res.ImageData = imageData

	// 10. 清理临时文件
	adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
	os.Remove(tempFileName)

	return res, nil
}
