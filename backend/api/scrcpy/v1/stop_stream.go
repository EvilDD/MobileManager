package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StopStreamReq 停止流请求
// @Description 停止设备流请求参数
type StopStreamReq struct {
	g.Meta   `path:"/stream/stop" tags:"串流管理" method:"post" summary:"停止串流"`
	DeviceId string `json:"deviceId" v:"required#设备ID不能为空" dc:"设备ID" example:"127.0.0.1:16480"`
}

// StopStreamRes 停止流响应
// @Description 停止设备流响应参数
type StopStreamRes struct{}
