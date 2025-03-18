package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"backend/internal/controller"
	"backend/internal/controller/hello"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
				)
			})

			// 添加设备相关路由
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				// 设备相关接口
				group.GET("/devices/list", controller.DeviceController.List)

				// 分组相关接口
				group.GET("/groups/list", controller.GroupController.List)
				group.POST("/groups/create", controller.GroupController.Create)
				group.PUT("/groups/update", controller.GroupController.Update)
				group.DELETE("/groups/delete", controller.GroupController.Delete)
			})

			s.Run()
			return nil
		},
	}
)
