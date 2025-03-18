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
				)
			})

			// 添加设备相关路由
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				// 添加中间件
				group.Middleware(ghttp.MiddlewareHandlerResponse)

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
