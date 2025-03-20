package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StartStreamReq 启动流请求
// @Description 启动设备流请求参数
type StartStreamReq struct {
	g.Meta   `path:"/stream/start" tags:"串流管理" method:"post" summary:"开始串流"`
	DeviceId string `json:"deviceId" v:"required#设备ID不能为空" dc:"设备ID" example:"127.0.0.1:16480"`
}

// StartStreamRes 启动流响应
// @Description 启动设备流响应参数
type StartStreamRes struct {
	Port int    `json:"port"    dc:"WebSocket端口号" example:"8886"`
	Url  string `json:"url"     dc:"WebSocket连接URL" example:"ws://localhost:8886"`
}
