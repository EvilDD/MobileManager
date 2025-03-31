package app

import (
	"context"
	"strings"

	v1 "backend/api/app/v1"
	"backend/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// List 获取应用列表
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	return service.AppService.List(ctx, req)
}

// Create 创建应用
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	if err = service.AppService.Create(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CreateRes{}, nil
}

// Delete 删除应用
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = service.AppService.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}

// Install 安装应用
func (c *ControllerV1) Install(ctx context.Context, req *v1.InstallReq) (res *v1.InstallRes, err error) {
	if err = service.AppService.Install(ctx, req); err != nil {
		return nil, err
	}
	return &v1.InstallRes{}, nil
}

// Uninstall 卸载应用
func (c *ControllerV1) Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.UninstallRes, err error) {
	if err = service.AppService.Uninstall(ctx, req); err != nil {
		return nil, err
	}
	return &v1.UninstallRes{}, nil
}

// Start 启动应用
func (c *ControllerV1) Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error) {
	if err = service.AppService.Start(ctx, req); err != nil {
		return nil, err
	}
	return &v1.StartRes{}, nil
}

// Upload 上传APK文件
func (c *ControllerV1) Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error) {
	// 日志记录
	g.Log().Debug(ctx, "开始处理文件上传请求")

	// 从请求上下文中获取上传文件
	r := g.RequestFromCtx(ctx)
	if r == nil {
		g.Log().Error(ctx, "获取HTTP请求对象失败")
		return nil, gerror.New("无法处理上传请求")
	}

	// 手动解析multipart表单，设置足够大的缓冲区
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		g.Log().Debug(ctx, "尝试解析multipart/form-data请求")
		if err := r.ParseMultipartForm(500 << 20); err != nil { // 500MB
			g.Log().Error(ctx, "解析文件上传表单失败:", err)
			return nil, gerror.Newf("解析文件上传表单失败: %v", err)
		}
	}

	// 记录Content-Type和其他请求头
	contentType := r.Header.Get("Content-Type")
	g.Log().Debug(ctx, "请求Content-Type:", contentType)
	g.Log().Debug(ctx, "请求方法:", r.Method)

	// 尝试直接获取上传文件
	file := r.GetUploadFile("file")
	if file != nil {
		g.Log().Debug(ctx, "控制器中直接找到上传文件:", file.Filename)
		req.File = file
	} else {
		g.Log().Warning(ctx, "控制器中未找到上传文件，将由服务层处理")
	}

	// 记录上传请求信息
	g.Log().Debug(ctx, "上传请求参数:", r.GetMap())

	// 记录请求体内容
	bodyStr := r.GetBodyString()
	if len(bodyStr) > 0 {
		// 如果内容太大，只记录前1000个字符
		if len(bodyStr) > 1000 {
			g.Log().Debug(ctx, "请求体(前1000字符):", bodyStr[:1000])
		} else {
			g.Log().Debug(ctx, "请求体:", bodyStr)
		}
	}

	// 将处理委托给Service层
	return service.AppService.Upload(ctx, req)
}

// BatchInstall 批量安装应用
func (c *ControllerV1) BatchInstall(ctx context.Context, req *v1.BatchInstallReq) (res *v1.BatchInstallRes, err error) {
	// 转换请求参数
	serviceReq := &service.BatchOperationReq{
		ID:        uint(req.Id),
		GroupID:   uint(req.GroupId),
		MaxWorker: req.MaxWorker,
	}
	return service.AppService.BatchInstall(ctx, serviceReq)
}

// BatchUninstall 批量卸载应用
func (c *ControllerV1) BatchUninstall(ctx context.Context, req *v1.BatchUninstallReq) (res *v1.BatchUninstallRes, err error) {
	// 转换请求参数
	serviceReq := &service.BatchOperationReq{
		ID:        uint(req.Id),
		GroupID:   uint(req.GroupId),
		MaxWorker: req.MaxWorker,
	}
	return service.AppService.BatchUninstall(ctx, serviceReq)
}

// BatchStart 批量启动应用
func (c *ControllerV1) BatchStart(ctx context.Context, req *v1.BatchStartReq) (res *v1.BatchStartRes, err error) {
	// 转换请求参数
	serviceReq := &service.BatchOperationReq{
		ID:        uint(req.Id),
		GroupID:   uint(req.GroupId),
		MaxWorker: req.MaxWorker,
	}
	return service.AppService.BatchStart(ctx, serviceReq)
}

// BatchTaskStatus 查询批量操作任务状态
func (c *ControllerV1) BatchTaskStatus(ctx context.Context, req *v1.BatchTaskStatusReq) (res *v1.BatchTaskStatusRes, err error) {
	return service.AppService.BatchTaskStatus(ctx, req)
}
