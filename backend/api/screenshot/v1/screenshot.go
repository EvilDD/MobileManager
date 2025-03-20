package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ScreenshotReq 截图请求
type ScreenshotReq struct {
	g.Meta   `path:"/screenshot/capture" tags:"截图管理" method:"post" summary:"批量设备截图"`
	DeviceId string `json:"deviceId" v:"required#设备ID不能为空" dc:"设备ID"`
	Quality  int    `json:"quality" v:"between:1,100#图片质量必须在1-100之间" dc:"图片质量(1-100)" d:"80"`
}

// // 打印请求参数
// func (r *ScreenshotReq) Print() {
// 	g.Log().Infof(context.Background(), "截图请求参数: deviceIds=%v, quality=%d", r.DeviceIds, r.Quality)
// }

// ScreenshotRes 截图响应
type ScreenshotRes struct {
	DeviceId  string `json:"deviceId" dc:"设备ID"`
	Success   bool   `json:"success" dc:"是否成功"`
	ImageData string `json:"imageData,omitempty" dc:"Base64编码的图片数据"`
	Error     string `json:"error,omitempty" dc:"错误信息"`
}
