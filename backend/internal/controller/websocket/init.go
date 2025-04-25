package websocket

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// 初始化日志配置
func init() {
	ctx := gctx.New()

	// 获取全局日志配置并应用到glog
	logger := g.Log()

	// 将glog全局配置与g.Log()同步
	glog.SetConfig(logger.GetConfig())

	// 设置glog的Writer为g.Log()的Writer
	glog.SetWriter(logger.GetWriter())

	g.Log().Debug(ctx, "websocket包日志初始化完成，已同步全局日志配置")
}
