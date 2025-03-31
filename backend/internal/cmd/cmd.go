package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	_ "backend/docs"
	"backend/internal/controller/app"
	"backend/internal/controller/device"
	groupctl "backend/internal/controller/group"
	"backend/internal/controller/hello"
	"backend/internal/controller/scrcpy"
	"backend/internal/controller/screenshot"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// 设置服务器配置参数
			s.SetMaxHeaderBytes(10 * 1024 * 1024)     // 10MB的请求头
			s.SetClientMaxBodySize(500 * 1024 * 1024) // 500MB的请求体大小限制

			// 添加CORS中间件
			s.Use(ghttp.MiddlewareCORS)

			// 设置OpenAPI路径
			s.SetOpenApiPath("/api.json")

			// 设置Swagger UI路径
			s.SetSwaggerPath("/swagger")

			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
					device.NewV1(),
					groupctl.NewV1(),
					scrcpy.NewV1(),
					screenshot.NewV1(),
					app.NewV1(),
				)
			})

			s.Run()
			return nil
		},
	}
)
