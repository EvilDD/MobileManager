package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ScreenshotReq 截图请求
type ScreenshotReq struct {
	g.Meta    `path:"/screenshot/capture" tags:"截图管理" method:"post" summary:"批量设备截图"`
	DeviceIds []string `json:"deviceIds" v:"required#设备ID列表不能为空" dc:"设备ID列表"`
	Quality   int      `json:"quality,optional" v:"min:1,max:100#图片质量必须在1-100之间" dc:"图片质量(1-100)" d:"80"`
}

// ScreenshotRes 截图响应
type ScreenshotRes struct {
	Screenshots []DeviceScreenshot `json:"screenshots" dc:"设备截图结果列表"`
}

// DeviceScreenshot 设备截图结果
type DeviceScreenshot struct {
	DeviceId  string `json:"deviceId" dc:"设备ID"`
	Success   bool   `json:"success" dc:"是否成功"`
	ImageUrl  string `json:"imageUrl,omitempty" dc:"截图URL地址(已弃用)"`
	ImageData string `json:"imageData,omitempty" dc:"Base64编码的图片数据"`
	Error     string `json:"error,omitempty" dc:"错误信息"`
}
