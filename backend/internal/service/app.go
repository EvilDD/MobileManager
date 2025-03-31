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

	// 解析APK文件
	apkInfo, err := apk.ParseAPK(filePath)
	if err != nil {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.Newf("解析APK文件失败: %v", err)
	}

	g.Log().Debug(ctx, "APK解析结果: 应用名=%s, 包名=%s, 版本=%s",
		apkInfo.ApplicationName, apkInfo.PackageName, apkInfo.VersionName)

	// 检查应用包名和版本是否已存在
	count, err := dao.App.Ctx(ctx).
		Where("package_name", apkInfo.PackageName).
		Where("version", apkInfo.VersionName).
		Count()
	if err != nil {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, err
	}
	if count > 0 {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.Newf("应用 %s 版本 %s 已存在", apkInfo.PackageName, apkInfo.VersionName)
	}

	// 创建应用记录
	_, err = dao.App.Ctx(ctx).Data(g.Map{
		"name":         apkInfo.ApplicationName,
		"package_name": apkInfo.PackageName,
		"version":      apkInfo.VersionName,
		"size":         written,
		"app_type":     "用户应用",
		"apk_path":     filePath,
	}).Insert()

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

// 任务管理器
var taskManager = struct {
	tasks map[string]*BatchTask
}{
	tasks: make(map[string]*BatchTask),
}

// BatchTask 批量操作任务
type BatchTask struct {
	TaskId    string
	Status    string
	Total     int
	Completed int
	Failed    int
	Results   []v1.BatchTaskResult
}

// BatchOperationReq 批量操作请求
type BatchOperationReq struct {
	ID        uint `json:"id"`
	GroupID   uint `json:"groupId"`
	MaxWorker int  `json:"maxWorker"`
}

// 验证并调整并发数
func (req *BatchOperationReq) validateAndAdjustMaxWorker(ctx context.Context) {
	// 从配置中获取最大并发数限制
	maxAllowedWorker := g.Cfg().MustGet(ctx, "batch.maxWorker").Int()
	if maxAllowedWorker <= 0 {
		maxAllowedWorker = 20 // 默认值
	}

	// 如果请求的并发数超过限制，则使用最大允许值
	if req.MaxWorker > maxAllowedWorker {
		g.Log().Noticef(ctx, "请求的并发数 %d 超过系统限制 %d，已自动调整", req.MaxWorker, maxAllowedWorker)
		req.MaxWorker = maxAllowedWorker
	}

	// 确保最小并发数为 1
	if req.MaxWorker < 1 {
		req.MaxWorker = 1
	}
}

// BatchInstall 批量安装应用
func (s *appService) BatchInstall(ctx context.Context, req *BatchOperationReq) (res *v1.BatchInstallRes, err error) {
	// 验证并调整并发数
	req.validateAndAdjustMaxWorker(ctx)

	// 查询应用信息
	var app *model.App
	err = dao.App.Ctx(ctx).Where("id", req.ID).Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, gerror.New("应用不存在")
	}

	// 检查APK文件是否存在
	if !gfile.Exists(app.ApkPath) {
		return nil, gerror.New("APK文件不存在")
	}

	// 获取分组设备列表
	var devices []*model.Device
	err = dao.Device.Ctx(ctx).Where("group_id", req.GroupID).Scan(&devices)
	if err != nil {
		return nil, err
	}

	// 如果分组下没有设备，返回空结果
	if len(devices) == 0 {
		taskId := grand.S(32)
		return &v1.BatchInstallRes{
			TaskId:    taskId,
			Total:     0,
			DeviceIds: []string{},
		}, nil
	}

	// 创建任务
	taskId := grand.S(32)
	deviceIds := make([]string, 0, len(devices))
	for _, device := range devices {
		deviceIds = append(deviceIds, device.DeviceId)
	}

	task := &BatchTask{
		TaskId:    taskId,
		Status:    v1.TaskStatusPending,
		Total:     len(devices),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0, len(devices)),
	}
	taskManager.tasks[taskId] = task

	// 启动后台任务
	go s.processBatchInstall(task, app, deviceIds, req.MaxWorker)

	return &v1.BatchInstallRes{
		TaskId:    taskId,
		Total:     len(devices),
		DeviceIds: deviceIds,
	}, nil
}

// BatchUninstall 批量卸载应用
func (s *appService) BatchUninstall(ctx context.Context, req *BatchOperationReq) (res *v1.BatchUninstallRes, err error) {
	// 验证并调整并发数
	req.validateAndAdjustMaxWorker(ctx)

	// 查询应用信息
	var app *model.App
	err = dao.App.Ctx(ctx).Where("id", req.ID).Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, gerror.New("应用不存在")
	}

	// 获取分组设备列表
	var devices []*model.Device
	err = dao.Device.Ctx(ctx).Where("group_id", req.GroupID).Scan(&devices)
	if err != nil {
		return nil, err
	}

	// 如果分组下没有设备，返回空结果
	if len(devices) == 0 {
		taskId := grand.S(32)
		return &v1.BatchUninstallRes{
			TaskId:    taskId,
			Total:     0,
			DeviceIds: []string{},
		}, nil
	}

	// 创建任务
	taskId := grand.S(32)
	deviceIds := make([]string, 0, len(devices))
	for _, device := range devices {
		deviceIds = append(deviceIds, device.DeviceId)
	}

	task := &BatchTask{
		TaskId:    taskId,
		Status:    v1.TaskStatusPending,
		Total:     len(devices),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0, len(devices)),
	}
	taskManager.tasks[taskId] = task

	// 启动后台任务
	go s.processBatchUninstall(task, app, deviceIds, req.MaxWorker)

	return &v1.BatchUninstallRes{
		TaskId:    taskId,
		Total:     len(devices),
		DeviceIds: deviceIds,
	}, nil
}

// BatchStart 批量启动应用
func (s *appService) BatchStart(ctx context.Context, req *BatchOperationReq) (res *v1.BatchStartRes, err error) {
	// 验证并调整并发数
	req.validateAndAdjustMaxWorker(ctx)

	// 查询应用信息
	var app *model.App
	err = dao.App.Ctx(ctx).Where("id", req.ID).Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, gerror.New("应用不存在")
	}

	// 获取分组设备列表
	var devices []*model.Device
	err = dao.Device.Ctx(ctx).Where("group_id", req.GroupID).Scan(&devices)
	if err != nil {
		return nil, err
	}

	// 如果分组下没有设备，返回空结果
	if len(devices) == 0 {
		taskId := grand.S(32)
		return &v1.BatchStartRes{
			TaskId:    taskId,
			Total:     0,
			DeviceIds: []string{},
		}, nil
	}

	// 创建任务
	taskId := grand.S(32)
	deviceIds := make([]string, 0, len(devices))
	for _, device := range devices {
		deviceIds = append(deviceIds, device.DeviceId)
	}

	task := &BatchTask{
		TaskId:    taskId,
		Status:    v1.TaskStatusPending,
		Total:     len(devices),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0, len(devices)),
	}
	taskManager.tasks[taskId] = task

	// 启动后台任务
	go s.processBatchStart(task, app, deviceIds, req.MaxWorker)

	return &v1.BatchStartRes{
		TaskId:    taskId,
		Total:     len(devices),
		DeviceIds: deviceIds,
	}, nil
}

// BatchTaskStatus 查询批量操作任务状态
func (s *appService) BatchTaskStatus(ctx context.Context, req *v1.BatchTaskStatusReq) (res *v1.BatchTaskStatusRes, err error) {
	task, ok := taskManager.tasks[req.TaskId]
	if !ok {
		return nil, gerror.New("任务不存在")
	}

	return &v1.BatchTaskStatusRes{
		TaskId:    task.TaskId,
		Status:    task.Status,
		Total:     task.Total,
		Completed: task.Completed,
		Failed:    task.Failed,
		Results:   task.Results,
	}, nil
}

// processBatchInstall 处理批量安装任务
func (s *appService) processBatchInstall(task *BatchTask, app *model.App, deviceIds []string, maxWorker int) {
	task.Status = v1.TaskStatusRunning

	// 创建工作池
	workerChan := make(chan struct{}, maxWorker)
	doneChan := make(chan struct{})

	// 启动工作协程
	for _, deviceId := range deviceIds {
		workerChan <- struct{}{} // 获取工作槽
		go func(deviceId string) {
			defer func() {
				<-workerChan // 释放工作槽
				if task.Completed+task.Failed == task.Total {
					task.Status = v1.TaskStatusComplete
					close(doneChan)
				}
			}()

			// 执行安装
			output, err := adb.InstallApp(deviceId, app.ApkPath)
			result := v1.BatchTaskResult{
				DeviceId: deviceId,
				Status:   v1.TaskStatusComplete,
			}

			if err != nil {
				result.Status = v1.TaskStatusFailed
				result.Message = err.Error() + ": " + output
				task.Failed++
			} else {
				result.Message = "安装成功"
				task.Completed++
			}

			task.Results = append(task.Results, result)
		}(deviceId)
	}

	// 等待所有任务完成
	<-doneChan
}

// processBatchUninstall 处理批量卸载任务
func (s *appService) processBatchUninstall(task *BatchTask, app *model.App, deviceIds []string, maxWorker int) {
	task.Status = v1.TaskStatusRunning

	// 创建工作池
	workerChan := make(chan struct{}, maxWorker)
	doneChan := make(chan struct{})

	// 启动工作协程
	for _, deviceId := range deviceIds {
		workerChan <- struct{}{} // 获取工作槽
		go func(deviceId string) {
			defer func() {
				<-workerChan // 释放工作槽
				if task.Completed+task.Failed == task.Total {
					task.Status = v1.TaskStatusComplete
					close(doneChan)
				}
			}()

			// 执行卸载
			output, err := adb.UninstallApp(deviceId, app.PackageName)
			result := v1.BatchTaskResult{
				DeviceId: deviceId,
				Status:   v1.TaskStatusComplete,
			}

			if err != nil {
				result.Status = v1.TaskStatusFailed
				result.Message = err.Error() + ": " + output
				task.Failed++
			} else {
				result.Message = "卸载成功"
				task.Completed++
			}

			task.Results = append(task.Results, result)
		}(deviceId)
	}

	// 等待所有任务完成
	<-doneChan
}

// processBatchStart 处理批量启动任务
func (s *appService) processBatchStart(task *BatchTask, app *model.App, deviceIds []string, maxWorker int) {
	task.Status = v1.TaskStatusRunning

	// 创建工作池
	workerChan := make(chan struct{}, maxWorker)
	doneChan := make(chan struct{})

	// 启动工作协程
	for _, deviceId := range deviceIds {
		workerChan <- struct{}{} // 获取工作槽
		go func(deviceId string) {
			defer func() {
				<-workerChan // 释放工作槽
				if task.Completed+task.Failed == task.Total {
					task.Status = v1.TaskStatusComplete
					close(doneChan)
				}
			}()

			// 执行启动
			output, err := adb.StartApp(deviceId, app.PackageName)
			result := v1.BatchTaskResult{
				DeviceId: deviceId,
				Status:   v1.TaskStatusComplete,
			}

			if err != nil {
				result.Status = v1.TaskStatusFailed
				result.Message = err.Error() + ": " + output
				task.Failed++
			} else {
				result.Message = "启动成功"
				task.Completed++
			}

			task.Results = append(task.Results, result)
		}(deviceId)
	}

	// 等待所有任务完成
	<-doneChan
}

// BatchInstallByDevices 按设备ID批量安装应用
func (s *appService) BatchInstallByDevices(ctx context.Context, req *v1.BatchInstallByDevicesReq) (res *v1.BatchInstallByDevicesRes, err error) {
	// 验证并调整并发数
	serviceReq := &BatchOperationReq{
		ID:        uint(req.Id),
		MaxWorker: req.MaxWorker,
	}
	serviceReq.validateAndAdjustMaxWorker(ctx)

	// 查询应用信息
	var app *model.App
	err = dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, gerror.New("应用不存在")
	}

	// 检查APK文件是否存在
	if !gfile.Exists(app.ApkPath) {
		return nil, gerror.New("APK文件不存在")
	}

	// 创建任务
	taskId := grand.S(32)
	task := &BatchTask{
		TaskId:    taskId,
		Status:    v1.TaskStatusPending,
		Total:     len(req.DeviceIds),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0, len(req.DeviceIds)),
	}
	taskManager.tasks[taskId] = task

	// 启动后台任务
	go s.processBatchInstall(task, app, req.DeviceIds, req.MaxWorker)

	return &v1.BatchInstallByDevicesRes{
		TaskId:    taskId,
		Total:     len(req.DeviceIds),
		DeviceIds: req.DeviceIds,
	}, nil
}

// BatchUninstallByDevices 按设备ID批量卸载应用
func (s *appService) BatchUninstallByDevices(ctx context.Context, req *v1.BatchUninstallByDevicesReq) (res *v1.BatchUninstallByDevicesRes, err error) {
	// 验证并调整并发数
	serviceReq := &BatchOperationReq{
		ID:        uint(req.Id),
		MaxWorker: req.MaxWorker,
	}
	serviceReq.validateAndAdjustMaxWorker(ctx)

	// 查询应用信息
	var app *model.App
	err = dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, gerror.New("应用不存在")
	}

	// 创建任务
	taskId := grand.S(32)
	task := &BatchTask{
		TaskId:    taskId,
		Status:    v1.TaskStatusPending,
		Total:     len(req.DeviceIds),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0, len(req.DeviceIds)),
	}
	taskManager.tasks[taskId] = task

	// 启动后台任务
	go s.processBatchUninstall(task, app, req.DeviceIds, req.MaxWorker)

	return &v1.BatchUninstallByDevicesRes{
		TaskId:    taskId,
		Total:     len(req.DeviceIds),
		DeviceIds: req.DeviceIds,
	}, nil
}

// BatchStartByDevices 按设备ID批量启动应用
func (s *appService) BatchStartByDevices(ctx context.Context, req *v1.BatchStartByDevicesReq) (res *v1.BatchStartByDevicesRes, err error) {
	// 验证并调整并发数
	serviceReq := &BatchOperationReq{
		ID:        uint(req.Id),
		MaxWorker: req.MaxWorker,
	}
	serviceReq.validateAndAdjustMaxWorker(ctx)

	// 查询应用信息
	var app *model.App
	err = dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, gerror.New("应用不存在")
	}

	// 创建任务
	taskId := grand.S(32)
	task := &BatchTask{
		TaskId:    taskId,
		Status:    v1.TaskStatusPending,
		Total:     len(req.DeviceIds),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0, len(req.DeviceIds)),
	}
	taskManager.tasks[taskId] = task

	// 启动后台任务
	go s.processBatchStart(task, app, req.DeviceIds, req.MaxWorker)

	return &v1.BatchStartByDevicesRes{
		TaskId:    taskId,
		Total:     len(req.DeviceIds),
		DeviceIds: req.DeviceIds,
	}, nil
}

// processBatchStop 处理批量停止任务
func (s *appService) processBatchStop(task *BatchTask, app *model.App, deviceIds []string, maxWorker int) {
	task.Status = v1.TaskStatusRunning

	// 创建工作池
	workerChan := make(chan struct{}, maxWorker)
	doneChan := make(chan struct{})

	// 启动工作协程
	for _, deviceId := range deviceIds {
		workerChan <- struct{}{} // 获取工作槽
		go func(deviceId string) {
			defer func() {
				<-workerChan // 释放工作槽
				if task.Completed+task.Failed == task.Total {
					task.Status = v1.TaskStatusComplete
					close(doneChan)
				}
			}()

			// 执行停止
			output, err := adb.StopApp(deviceId, app.PackageName)
			result := v1.BatchTaskResult{
				DeviceId: deviceId,
				Status:   v1.TaskStatusComplete,
			}

			if err != nil {
				result.Status = v1.TaskStatusFailed
				result.Message = err.Error() + ": " + output
				task.Failed++
			} else {
				result.Message = "停止成功"
				task.Completed++
			}

			task.Results = append(task.Results, result)
		}(deviceId)
	}

	// 等待所有任务完成
	<-doneChan
}

// BatchStopByDevices 按设备ID批量停止应用
func (s *appService) BatchStopByDevices(ctx context.Context, req *v1.BatchStopByDevicesReq) (res *v1.BatchStopByDevicesRes, err error) {
	// 验证并调整并发数
	serviceReq := &BatchOperationReq{
		ID:        uint(req.Id),
		MaxWorker: req.MaxWorker,
	}
	serviceReq.validateAndAdjustMaxWorker(ctx)

	// 查询应用信息
	var app *model.App
	err = dao.App.Ctx(ctx).Where("id", req.Id).Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, gerror.New("应用不存在")
	}

	// 创建任务
	taskId := grand.S(32)
	task := &BatchTask{
		TaskId:    taskId,
		Status:    v1.TaskStatusPending,
		Total:     len(req.DeviceIds),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0, len(req.DeviceIds)),
	}
	taskManager.tasks[taskId] = task

	// 启动后台任务
	go s.processBatchStop(task, app, req.DeviceIds, req.MaxWorker)

	return &v1.BatchStopByDevicesRes{
		TaskId:    taskId,
		Total:     len(req.DeviceIds),
		DeviceIds: req.DeviceIds,
	}, nil
}
