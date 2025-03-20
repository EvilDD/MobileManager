package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	v1 "backend/api/screenshot/v1"
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

// ... existing code ...
func (s *screenshotService) Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error) {
	deviceId := req.DeviceId
	res = &v1.ScreenshotRes{
		DeviceId:  deviceId,
		Success:   false,
		ImageData: "",
		Error:     "",
	}

	// 生成唯一的临时文件名
	timestamp := time.Now().UnixNano()
	tempFileName := fmt.Sprintf("screenshot_%s_%d.png", strings.Replace(deviceId, ":", "_", -1), timestamp)
	deviceTempPath := fmt.Sprintf("/data/local/tmp/%s", tempFileName)

	// 1. 在设备上执行截图命令
	cmdScreencap := exec.Command("adb", "-s", deviceId, "shell", "screencap", "-p", deviceTempPath)
	if err = cmdScreencap.Run(); err != nil {
		res.Error = fmt.Sprintf("设备截图失败: %v", err)
		return
	}

	// 2. 从设备中拉取截图文件
	var buf bytes.Buffer
	cmdPull := exec.Command("adb", "-s", deviceId, "pull", deviceTempPath)
	cmdPull.Stdout = &buf
	if err = cmdPull.Run(); err != nil {
		res.Error = fmt.Sprintf("拉取截图失败: %v", err)
		// 清理设备上的临时文件
		exec.Command("adb", "-s", deviceId, "shell", "rm", deviceTempPath).Run()
		return
	}

	// 3. 读取本地文件
	imgData, err := os.ReadFile(tempFileName)
	if err != nil {
		res.Error = fmt.Sprintf("读取截图文件失败: %v", err)
		// 清理设备上和本地的临时文件
		exec.Command("adb", "-s", deviceId, "shell", "rm", deviceTempPath).Run()
		os.Remove(tempFileName)
		return
	}

	// 4. 解码PNG图片
	img, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		res.Error = fmt.Sprintf("解码PNG图片失败: %v", err)
		// 清理临时文件
		exec.Command("adb", "-s", deviceId, "shell", "rm", deviceTempPath).Run()
		os.Remove(tempFileName)
		return
	}

	// 5. 压缩为JPEG并编码为Base64
	var jpegBuf bytes.Buffer
	quality := req.Quality
	if quality == 0 {
		quality = 80 // 默认质量
	}

	err = jpeg.Encode(&jpegBuf, img, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		res.Error = fmt.Sprintf("JPEG编码失败: %v", err)
		// 清理临时文件
		exec.Command("adb", "-s", deviceId, "shell", "rm", deviceTempPath).Run()
		os.Remove(tempFileName)
		return
	}

	// 6. 转换为Base64
	res.Success = true
	res.ImageData = fmt.Sprintf("data:image/jpeg;base64,%s",
		base64.StdEncoding.EncodeToString(jpegBuf.Bytes()))

	// 7. 清理临时文件
	exec.Command("adb", "-s", deviceId, "shell", "rm", deviceTempPath).Run()
	os.Remove(tempFileName)
	return res, nil
}
