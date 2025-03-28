package service

import (
	"context"
	"os"
	"strings"
	"time"

	v1 "backend/api/app/v1"
	"backend/internal/dao"
	"backend/internal/model"
	"backend/utility/adb"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/grand"
)

var AppService = appService{}

type appService struct{}

// List 获取应用列表
func (s *appService) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	res = &v1.ListRes{
		List:     make([]v1.App, 0),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	m := dao.App.Ctx(ctx)

	// 应用类型筛选
	if req.AppType != "" {
		m = m.Where("app_type", req.AppType)
	}

	// 关键词搜索
	if req.Keyword != "" {
		m = m.WhereLike("name", "%"+req.Keyword+"%").
			WhereOrLike("package_name", "%"+req.Keyword+"%")
	}

	// 获取总数
	res.Total, err = m.Count()
	if err != nil {
		return nil, err
	}

	// 获取分页数据
	var apps []*model.App
	err = m.Page(req.Page, req.PageSize).Scan(&apps)
	if err != nil {
		return nil, err
	}

	// 转换为API返回结构
	for _, app := range apps {
		res.List = append(res.List, v1.App{
			Id:          app.Id,
			Name:        app.Name,
			PackageName: app.PackageName,
			Version:     app.Version,
			Size:        app.Size,
			AppType:     app.AppType,
			ApkPath:     app.ApkPath,
			CreatedAt:   app.CreatedAt.String(),
			UpdatedAt:   app.UpdatedAt.String(),
		})
	}

	return
}

// Create 创建应用
func (s *appService) Create(ctx context.Context, req *v1.CreateReq) error {
	// 检查应用包名是否已存在
	count, err := dao.App.Ctx(ctx).Where("package_name", req.PackageName).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("应用包名已存在")
	}

	// 检查APK文件是否存在
	if !gfile.Exists(req.ApkPath) {
		return gerror.New("APK文件不存在")
	}

	// 创建应用记录
	_, err = dao.App.Ctx(ctx).Data(g.Map{
		"name":         req.Name,
		"package_name": req.PackageName,
		"version":      req.Version,
		"size":         req.Size,
		"app_type":     req.AppType,
		"apk_path":     req.ApkPath,
	}).Insert()

	return err
}

// Delete 删除应用
func (s *appService) Delete(ctx context.Context, req *v1.DeleteReq) error {
	// 查询应用信息
	var app *model.App
	err := dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return err
	}
	if app == nil {
		return gerror.New("应用不存在")
	}

	// 如果APK文件存在，则删除
	if gfile.Exists(app.ApkPath) {
		err = gfile.Remove(app.ApkPath)
		if err != nil {
			return gerror.Newf("删除APK文件失败: %v", err)
		}
	}

	// 从数据库中删除应用记录
	_, err = dao.App.Ctx(ctx).Where("id", req.Id).Delete()
	return err
}

// Install 安装应用
func (s *appService) Install(ctx context.Context, req *v1.InstallReq) error {
	// 查询应用信息
	var app *model.App
	err := dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return err
	}
	if app == nil {
		return gerror.New("应用不存在")
	}

	// 检查APK文件是否存在
	if !gfile.Exists(app.ApkPath) {
		return gerror.New("APK文件不存在")
	}

	// 检查设备是否存在
	var device *model.Device
	err = dao.Device.Ctx(ctx).Where("device_id", req.DeviceId).Scan(&device)
	if err != nil {
		return err
	}
	if device == nil {
		return gerror.New("设备不存在")
	}

	// 使用ADB安装应用
	output, err := adb.InstallApp(req.DeviceId, app.ApkPath)
	if err != nil {
		return gerror.Newf("安装应用失败: %v, 输出: %s", err, output)
	}

	return nil
}

// Uninstall 卸载应用
func (s *appService) Uninstall(ctx context.Context, req *v1.UninstallReq) error {
	// 查询应用信息
	var app *model.App
	err := dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return err
	}
	if app == nil {
		return gerror.New("应用不存在")
	}

	// 检查设备是否存在
	var device *model.Device
	err = dao.Device.Ctx(ctx).Where("device_id", req.DeviceId).Scan(&device)
	if err != nil {
		return err
	}
	if device == nil {
		return gerror.New("设备不存在")
	}

	// 使用ADB卸载应用
	output, err := adb.UninstallApp(req.DeviceId, app.PackageName)
	if err != nil {
		return gerror.Newf("卸载应用失败: %v, 输出: %s", err, output)
	}

	return nil
}

// Start 启动应用
func (s *appService) Start(ctx context.Context, req *v1.StartReq) error {
	// 查询应用信息
	var app *model.App
	err := dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return err
	}
	if app == nil {
		return gerror.New("应用不存在")
	}

	// 检查设备是否存在
	var device *model.Device
	err = dao.Device.Ctx(ctx).Where("device_id", req.DeviceId).Scan(&device)
	if err != nil {
		return err
	}
	if device == nil {
		return gerror.New("设备不存在")
	}

	// 使用ADB启动应用
	output, err := adb.StartApp(req.DeviceId, app.PackageName)
	if err != nil {
		return gerror.Newf("启动应用失败: %v, 输出: %s", err, output)
	}

	return nil
}

// Upload 上传APK文件
func (s *appService) Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error) {
	res = &v1.UploadRes{}

	// 详细日志记录
	g.Log().Debug(ctx, "Service层开始处理上传请求")

	// 从HTTP请求中直接获取文件
	r := g.RequestFromCtx(ctx)
	if r == nil {
		g.Log().Error(ctx, "获取HTTP请求对象失败")
		return nil, gerror.New("无法处理上传请求")
	}

	// 确保请求方法是POST
	g.Log().Debug(ctx, "请求方法:", r.Method)
	if r.Method != "POST" {
		g.Log().Error(ctx, "请求方法不正确:", r.Method)
		return nil, gerror.New("只支持POST方法上传文件")
	}

	// 记录Content-Type，用于调试
	contentType := r.Header.Get("Content-Type")
	g.Log().Debug(ctx, "请求Content-Type:", contentType)

	// 改进：更灵活地处理Content-Type
	var file *ghttp.UploadFile

	// 处理不同的Content-Type
	if strings.Contains(contentType, "multipart/form-data") {
		// 标准的multipart/form-data处理
		g.Log().Debug(ctx, "处理multipart/form-data请求")
		file = r.GetUploadFile("file")
	} else if strings.Contains(contentType, "application/json") {
		// 尝试从JSON请求中解析文件信息
		g.Log().Debug(ctx, "尝试从JSON请求中解析文件信息")

		// 先打印完整的请求体便于调试
		bodyContent := r.GetBodyString()
		g.Log().Debug(ctx, "JSON请求体内容:", bodyContent)

		// 查看请求的表单数据
		formData := r.GetFormMap()
		g.Log().Debug(ctx, "JSON请求表单数据:", formData)

		// 尝试手动解析multipart表单
		g.Log().Debug(ctx, "尝试手动解析multipart表单...")
		if err := r.ParseMultipartForm(500 << 20); err != nil { // 500MB 最大内存
			g.Log().Warning(ctx, "解析multipart表单失败:", err)
		} else {
			g.Log().Debug(ctx, "成功解析multipart表单")
		}
	} else {
		// 其他类型的Content-Type
		g.Log().Debug(ctx, "处理其他类型的Content-Type请求")
	}

	// 尝试获取上传的文件
	if file == nil {
		file = r.GetUploadFile("file")
	}

	// 尝试从多种来源获取文件
	if file == nil {
		g.Log().Debug(ctx, "直接获取文件失败，尝试从表单中获取")

		// 记录所有请求信息以便调试
		g.Log().Debug(ctx, "请求体内容:", r.GetBodyString())
		g.Log().Debug(ctx, "请求头:", r.Header)

		// 记录请求表单数据
		formData := r.GetFormMap()
		g.Log().Debug(ctx, "表单数据:", formData)

		// 如果有文件上传表单，尝试读取
		if r.MultipartForm != nil && r.MultipartForm.File != nil {
			g.Log().Debug(ctx, "找到表单文件字段数:", len(r.MultipartForm.File))

			for name, fileHeaders := range r.MultipartForm.File {
				g.Log().Debug(ctx, "表单字段:", name, "包含", len(fileHeaders), "个文件")
				if len(fileHeaders) > 0 && fileHeaders[0] != nil {
					// 尝试获取第一个文件
					fileInfo := fileHeaders[0]
					g.Log().Debug(ctx, "找到文件:", fileInfo.Filename, "大小:", fileInfo.Size)
					tempFile := r.GetUploadFile(name)
					if tempFile != nil {
						file = tempFile
						g.Log().Debug(ctx, "成功获取到上传文件:", tempFile.Filename)
						break
					}
				}
			}
		}

		// 如果还是没有找到文件，尝试获取所有可能的文件上传字段
		if file == nil {
			g.Log().Debug(ctx, "尝试查找所有可能的文件字段")
			possibleFields := []string{"file", "apk", "uploadFile", "apkFile"}
			for _, fieldName := range possibleFields {
				tempFile := r.GetUploadFile(fieldName)
				if tempFile != nil {
					file = tempFile
					g.Log().Debug(ctx, "在字段", fieldName, "中找到文件:", tempFile.Filename)
					break
				}
			}
		}

		// 如果还是没有找到文件
		if file == nil {
			g.Log().Error(ctx, "未找到上传文件")
			return nil, gerror.New("未找到上传文件")
		}
	}

	g.Log().Debug(ctx, "成功获取到上传文件:", file.Filename)

	// 检查文件扩展名
	filename := file.Filename
	if !strings.HasSuffix(strings.ToLower(filename), ".apk") {
		g.Log().Error(ctx, "文件扩展名不正确:", filename)
		return nil, gerror.New("只能上传APK文件")
	}

	// 确保保存目录存在
	uploadDir := "resource/apk"
	g.Log().Debug(ctx, "检查上传目录:", uploadDir)

	if !gfile.Exists(uploadDir) {
		g.Log().Debug(ctx, "上传目录不存在，准备创建")
		if err = gfile.Mkdir(uploadDir); err != nil {
			g.Log().Error(ctx, "创建目录失败:", err)
			return nil, gerror.Newf("创建目录失败: %v", err)
		}
		g.Log().Debug(ctx, "成功创建上传目录")
	}

	// 生成唯一文件名，使用当前时间 + 随机数
	now := time.Now().Format("20060102150405")
	random := gconv.String(grand.Intn(10000))
	uniqueFileName := now + "_" + random + ".apk"
	savePath := gfile.Join(uploadDir, uniqueFileName)

	g.Log().Debug(ctx, "准备保存文件到:", savePath)

	// 保存文件
	newFilename, err := file.Save(savePath)
	if err != nil {
		g.Log().Error(ctx, "保存文件失败:", err)
		return nil, gerror.Newf("保存文件失败: %v", err)
	}
	g.Log().Debug(ctx, "文件已保存，新文件名:", newFilename)

	// 确保文件已成功保存
	if !gfile.Exists(savePath) {
		g.Log().Error(ctx, "保存文件后无法找到文件:", savePath)
		return nil, gerror.New("文件保存失败，无法找到上传的文件")
	}

	// 获取文件大小
	fileInfo, err := os.Stat(savePath)
	if err != nil {
		g.Log().Error(ctx, "获取文件信息失败:", err)
		// 如果获取文件信息失败，尝试删除文件
		_ = gfile.Remove(savePath)
		return nil, gerror.Newf("获取文件信息失败: %v", err)
	}
	fileSize := fileInfo.Size() / 1024 // KB
	g.Log().Debug(ctx, "文件大小:", fileSize, "KB")

	// 设置返回信息
	res.FileName = filename
	res.FileSize = fileSize
	res.FilePath = savePath

	g.Log().Debug(ctx, "文件上传成功，返回信息:", res)
	return res, nil
}
