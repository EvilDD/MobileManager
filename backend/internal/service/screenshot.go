package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"os/exec"
	"sync"

	v1 "backend/api/screenshot/v1"

	"github.com/gogf/gf/v2/os/gfile"
)

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
		localScreenshot = &screenshotService{}
	})
	return localScreenshot
}

type screenshotService struct{}

// Capture 批量设备截图
func (s *screenshotService) Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error) {
	res = &v1.ScreenshotRes{
		Screenshots: make([]v1.DeviceScreenshot, 0, len(req.DeviceIds)),
	}

	// 创建截图保存目录
	screenshotDir := "public/screenshots"
	if err = gfile.Mkdir(screenshotDir); err != nil {
		return nil, fmt.Errorf("创建截图目录失败: %v", err)
	}

	// 使用信号量限制并发数
	semaphore := make(chan struct{}, 50)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, deviceId := range req.DeviceIds {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(deviceId string) {
			defer func() {
				<-semaphore // 释放信号量
				wg.Done()
			}()

			screenshot := v1.DeviceScreenshot{
				DeviceId: deviceId,
				Success:  false,
			}

			// 使用ADB命令通过管道直接获取截图数据
			cmd := exec.Command("adb", "-s", deviceId, "shell", "screencap", "-p")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				screenshot.Error = fmt.Sprintf("创建管道失败: %v", err)
				mu.Lock()
				res.Screenshots = append(res.Screenshots, screenshot)
				mu.Unlock()
				return
			}

			if err := cmd.Start(); err != nil {
				screenshot.Error = fmt.Sprintf("执行截图命令失败: %v", err)
				mu.Lock()
				res.Screenshots = append(res.Screenshots, screenshot)
				mu.Unlock()
				return
			}

			// 读取PNG图片数据
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, stdout); err != nil {
				screenshot.Error = fmt.Sprintf("读取截图数据失败: %v", err)
				mu.Lock()
				res.Screenshots = append(res.Screenshots, screenshot)
				mu.Unlock()
				return
			}

			if err := cmd.Wait(); err != nil {
				screenshot.Error = fmt.Sprintf("等待命令完成失败: %v", err)
				mu.Lock()
				res.Screenshots = append(res.Screenshots, screenshot)
				mu.Unlock()
				return
			}

			// 解码PNG图片
			img, err := png.Decode(bytes.NewReader(buf.Bytes()))
			if err != nil {
				screenshot.Error = fmt.Sprintf("解码PNG图片失败: %v", err)
				mu.Lock()
				res.Screenshots = append(res.Screenshots, screenshot)
				mu.Unlock()
				return
			}

			// 压缩为JPEG并编码为Base64
			var jpegBuf bytes.Buffer
			quality := req.Quality
			if quality == 0 {
				quality = 80 // 默认质量
			}

			err = jpeg.Encode(&jpegBuf, img, &jpeg.Options{
				Quality: quality,
			})
			if err != nil {
				screenshot.Error = fmt.Sprintf("JPEG编码失败: %v", err)
				mu.Lock()
				res.Screenshots = append(res.Screenshots, screenshot)
				mu.Unlock()
				return
			}

			// 转换为Base64
			screenshot.Success = true
			screenshot.ImageData = fmt.Sprintf("data:image/jpeg;base64,%s",
				base64.StdEncoding.EncodeToString(jpegBuf.Bytes()))

			// 线程安全地添加结果
			mu.Lock()
			res.Screenshots = append(res.Screenshots, screenshot)
			mu.Unlock()
		}(deviceId)
	}

	wg.Wait()
	return res, nil
}
