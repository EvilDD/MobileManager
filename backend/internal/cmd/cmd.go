package cmd

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"

	_ "backend/docs"
	"backend/internal/controller/device"
	groupctl "backend/internal/controller/group"
	"backend/internal/controller/hello"
	"backend/internal/controller/scrcpy"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// 添加CORS中间件
			s.Use(ghttp.MiddlewareCORS)

			// 添加swagger UI
			s.BindHandler("/swagger/*any", func(r *ghttp.Request) {
				handler := httpSwagger.Handler(
					httpSwagger.URL("/swagger/doc.json"), // 文档的URL
				)
				// 将http.Handler转换为ghttp.Handler
				handler.ServeHTTP(r.Response.Writer, r.Request)
			})

			// 添加swagger JSON文档
			s.BindHandler("/swagger/doc.json", func(r *ghttp.Request) {
				doc, err := swag.ReadDoc()
				if err != nil {
					r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
					return
				}
				r.Response.WriteJsonExit(doc)
			})

			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
					device.NewV1(),
					groupctl.NewV1(),
					scrcpy.NewV1(),
				)
			})

			s.Run()
			return nil
		},
	}
)
