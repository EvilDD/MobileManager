package boot

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

// 初始化日志配置
func init() {
	ctx := gctx.New()

	// 确保日志目录存在
	logPath := g.Cfg().MustGet(ctx, "logger.path", "./logs").String()
	if !gfile.Exists(logPath) {
		if err := gfile.Mkdir(logPath); err != nil {
			g.Log().Fatal(ctx, "创建日志目录失败:", err)
		}
	}

	// 大部分配置已在config.yaml中设置，这里仅确保目录存在
	g.Log().Info(ctx, "日志系统初始化完成，日志写入目录:", logPath)
}
