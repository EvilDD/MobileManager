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

func (s *screenshotService) Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error) {
	deviceId := req.DeviceId
	res = &v1.ScreenshotRes{
		DeviceId:  deviceId,
		Success:   false,
		ImageData: "",
		Error:     "",
	}

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
		adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
		os.Remove(tempFileName)
		return
	}

	// 7. 转换为Base64
	res.Success = true
	res.ImageData = fmt.Sprintf("data:image/jpeg;base64,%s",
		base64.StdEncoding.EncodeToString(jpegBuf.Bytes()))

	// 8. 清理临时文件
	adbTool.RemoveDeviceFile(deviceId, deviceTempPath)
	os.Remove(tempFileName)
	return res, nil
}
