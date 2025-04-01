package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"sync"
	"time"

	v1 "backend/api/screenshot/v1"
	"backend/utility/adb"

	"github.com/gogf/gf/v2/os/glog"
)

// 缓存项结构
type screenshotCacheItem struct {
	imageData string    // Base64编码的图片数据
	timestamp time.Time // 缓存时间
	quality   int       // 图片质量
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
func (s *screenshotService) getCachedScreenshot(deviceID string, quality int) *screenshotCacheItem {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	if item, exists := s.cache[deviceID]; exists {
		timeSinceCache := time.Since(item.timestamp)
		if timeSinceCache <= s.ttl && item.quality == quality {
			// glog.Info(context.Background(), fmt.Sprintf("设备[%s]命中缓存, 缓存时间: %.2f秒", deviceID, timeSinceCache.Seconds()))
			return item
		}
		if timeSinceCache > s.ttl {
			// glog.Info(context.Background(), fmt.Sprintf("设备[%s]缓存已过期, 缓存时间: %.2f秒", deviceID, timeSinceCache.Seconds()))
		}
		if item.quality != quality {
			// glog.Info(context.Background(), fmt.Sprintf("设备[%s]缓存质量不匹配, 缓存质量: %d, 请求质量: %d", deviceID, item.quality, quality))
		}
	} else {
		// glog.Info(context.Background(), fmt.Sprintf("设备[%s]无缓存", deviceID))
	}
	return nil
}

// 设置缓存
func (s *screenshotService) setCacheScreenshot(deviceID string, imageData string, quality int) {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	s.cache[deviceID] = &screenshotCacheItem{
		imageData: imageData,
		timestamp: time.Now(),
		quality:   quality,
	}
	// glog.Info(context.Background(), fmt.Sprintf("设备[%s]设置新缓存, 质量: %d", deviceID, quality))
}

// 获取或创建进行中的请求标记
func (s *screenshotService) getOrCreateInProgress(deviceID string) (chan struct{}, bool) {
	s.inProgressMux.Lock()
	defer s.inProgressMux.Unlock()

	if ch, exists := s.inProgress[deviceID]; exists {
		glog.Info(context.Background(), fmt.Sprintf("设备[%s]已有请求在处理中, 等待复用", deviceID))
		return ch, false
	}

	ch := make(chan struct{})
	s.inProgress[deviceID] = ch
	glog.Info(context.Background(), fmt.Sprintf("设备[%s]开始新的截图请求", deviceID))
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

func (s *screenshotService) Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error) {
	deviceId := req.DeviceId
	quality := req.Quality
	if quality == 0 {
		quality = 80 // 默认质量
	}

	startTime := time.Now()
	defer func() {
		glog.Info(ctx, fmt.Sprintf("设备[%s]截图请求完成, 耗时: %.2f毫秒, 成功: %v",
			deviceId, float64(time.Since(startTime).Milliseconds()), res.Success))
	}()

	res = &v1.ScreenshotRes{
		DeviceId:  deviceId,
		Success:   false,
		ImageData: "",
		Error:     "",
	}

	// 尝试从缓存获取
	if cached := s.getCachedScreenshot(deviceId, quality); cached != nil {
		res.Success = true
		res.ImageData = cached.imageData
		return res, nil
	}

	// 检查是否有正在进行的请求
	ch, isFirst := s.getOrCreateInProgress(deviceId)
	if !isFirst {
		// 等待已有请求完成
		select {
		case <-ch:
			// 其他请求已完成，尝试从缓存获取
			if cached := s.getCachedScreenshot(deviceId, quality); cached != nil {
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
		return
	}

	// 生成唯一的临时文件名
	timestamp := time.Now().UnixNano()
	tempFileName := fmt.Sprintf("screenshot_%s_%d.png", strings.Replace(deviceId, ":", "_", -1), timestamp)
	deviceTempPath := fmt.Sprintf("/data/local/tmp/%s", tempFileName)

	// 2. 在设备上执行截图命令
	if err = adbTool.Screencap(deviceId, deviceTempPath); err != nil {
		res.Error = fmt.Sprintf("设备截图失败: %v", err)
		return
	}

	// 3. 从设备中拉取截图文件
	if err = adbTool.PullFile(deviceId, deviceTempPath, tempFileName); err != nil {
		res.Error = fmt.Sprintf("拉取截图失败: %v", err)
		// 清理设备上的临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		return
	}

	// 4. 读取本地文件
	imgData, err := os.ReadFile(tempFileName)
	if err != nil {
		res.Error = fmt.Sprintf("读取截图文件失败: %v", err)
		// 清理设备上和本地的临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return
	}

	// 5. 解码PNG图片
	img, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		res.Error = fmt.Sprintf("解码PNG图片失败: %v", err)
		// 清理临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return
	}

	// 6. 压缩为JPEG并编码为Base64
	var jpegBuf bytes.Buffer
	err = jpeg.Encode(&jpegBuf, img, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		res.Error = fmt.Sprintf("JPEG编码失败: %v", err)
		// 清理临时文件
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return
	}

	// 7. 转换为Base64
	imageData := fmt.Sprintf("data:image/jpeg;base64,%s",
		base64.StdEncoding.EncodeToString(jpegBuf.Bytes()))

	// 8. 设置缓存
	s.setCacheScreenshot(deviceId, imageData, quality)

	res.Success = true
	res.ImageData = imageData

	// 9. 清理临时文件
	adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
	os.Remove(tempFileName)
	return res, nil
}
