package service

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	v1 "backend/api/app/v1"
	"backend/internal/dao"
	"backend/internal/model"
	"backend/utility/adb"
	"backend/utility/apk"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
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

	file := req.File
	if file == nil {
		return nil, gerror.New("请选择要上传的APK文件")
	}

	// 检查文件类型
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".apk") {
		return nil, gerror.New("只能上传APK文件")
	}

	// 生成存储路径
	uploadPath := "resource/apk"
	if !gfile.Exists(uploadPath) {
		err = gfile.Mkdir(uploadPath)
		if err != nil {
			return nil, gerror.Newf("创建上传目录失败: %v", err)
		}
	}

	// 生成唯一文件名
	now := time.Now()
	uniqueName := now.Format("20060102150405") + "_" + grand.S(8) + ".apk"
	filePath := gfile.Join(uploadPath, uniqueName)

	// 创建目标文件
	dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, gerror.Newf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return nil, gerror.Newf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 复制文件内容
	written, err := io.Copy(dst, src)
	if err != nil {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.Newf("保存文件失败: %v", err)
	}

	// 确保写入的大小正确
	if written == 0 {
		// 删除空文件
		_ = gfile.Remove(filePath)
		return nil, gerror.New("保存的文件大小为0")
	}

	// 输出文件信息
	g.Log().Debugf(ctx, "文件已保存: %s, 大小: %d 字节", filePath, written)

	// 解析APK文件
	apkInfo, err := apk.ParseAPK(filePath)
	if err != nil {
		// 输出错误详情
		g.Log().Errorf(ctx, "解析APK文件失败: %v, 文件路径: %s", err, filePath)
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.Newf("解析APK文件失败: %v", err)
	}

	// 输出解析结果
	g.Log().Debugf(ctx, "APK解析结果: 应用名=%s, 包名=%s, 版本=%s",
		apkInfo.ApplicationName, apkInfo.PackageName, apkInfo.VersionName)

	// 如果解析结果为空，使用文件名作为默认值
	if apkInfo.ApplicationName == "" {
		apkInfo.ApplicationName = strings.TrimSuffix(file.Filename, ".apk")
	}
	if apkInfo.PackageName == "" {
		apkInfo.PackageName = "com.example." + strings.ToLower(strings.ReplaceAll(apkInfo.ApplicationName, " ", ""))
	}
	if apkInfo.VersionName == "" {
		apkInfo.VersionName = "1.0.0"
	}

	// 检查应用包名是否已存在
	count, err := dao.App.Ctx(ctx).Where("package_name", apkInfo.PackageName).Count()
	if err != nil {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, err
	}
	if count > 0 {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.New("应用包名已存在")
	}

	// 创建应用记录
	appData := g.Map{
		"name":         apkInfo.ApplicationName,
		"package_name": apkInfo.PackageName,
		"version":      apkInfo.VersionName,
		"size":         written,
		"app_type":     "用户应用",
		"apk_path":     filePath,
	}

	// 输出即将插入的数据
	g.Log().Debugf(ctx, "即将插入数据库的应用信息: %+v", appData)

	_, err = dao.App.Ctx(ctx).Data(appData).Insert()
	if err != nil {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.Newf("保存应用信息失败: %v", err)
	}

	res.FileName = file.Filename
	res.FileSize = written
	res.FilePath = filePath

	return
}
