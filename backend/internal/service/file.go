package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"strings"
	"sync"

	v1 "backend/api/file/v1"
	"backend/internal/dao"
	"backend/internal/model"
	"backend/utility/adb"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
)

// 文件服务
var FileService = fileService{}

// 文件服务结构体
type fileService struct{}

// 任务管理器
var fileTaskManager = struct {
	tasks map[string]*FileBatchTask
	mutex sync.Mutex
}{
	tasks: make(map[string]*FileBatchTask),
}

// FileBatchTask 批量操作任务
type FileBatchTask struct {
	TaskId    string
	Status    string
	Total     int
	Completed int
	Failed    int
	Results   []v1.BatchTaskResult
	Mutex     sync.Mutex
}

// 推送目标目录
const DeviceTargetDir = "/data/local/tmp"

// Upload 上传文件
func (s *fileService) Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error) {
	res = &v1.UploadRes{}

	file := req.File
	if file == nil {
		return nil, gerror.New("请选择要上传的文件")
	}

	// 获取文件扩展名并转换为小写
	fileExt := strings.ToLower(path.Ext(file.Filename))
	if fileExt == "" {
		fileExt = ".unknown"
	}

	// 获取MIME类型
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		// 尝试从扩展名推断MIME类型
		mimeType = mime.TypeByExtension(fileExt)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	// 获取文件类型（基于扩展名）
	fileType := s.getFileTypeByExt(fileExt)

	// 计算文件的MD5值
	src, err := file.Open()
	if err != nil {
		return nil, gerror.Newf("打开上传文件失败: %v", err)
	}

	hash := md5.New()
	written, err := io.Copy(hash, src)
	if err != nil {
		src.Close()
		return nil, gerror.Newf("计算文件MD5失败: %v", err)
	}
	src.Close()

	// 确保文件大小正确
	if written == 0 {
		return nil, gerror.New("文件大小为0")
	}

	md5Value := hex.EncodeToString(hash.Sum(nil))
	g.Log().Debugf(ctx, "文件MD5计算完成: %s, 大小: %d字节", md5Value, written)

	// 检查是否已存在相同MD5的文件
	var existingFile *model.File
	err = dao.File.Ctx(ctx).Where("md5", md5Value).Where("status", 1).Scan(&existingFile)
	if err != nil {
		g.Log().Errorf(ctx, "查询已存在文件失败: %v", err)
	}

	// 如果已存在相同MD5的文件，直接返回该文件信息
	if existingFile != nil {
		g.Log().Infof(ctx, "文件已存在，MD5: %s, 文件ID: %d", md5Value, existingFile.Id)

		res.FileId = existingFile.Id
		res.FileName = existingFile.Name
		res.FileSize = existingFile.FileSize
		res.FilePath = existingFile.FilePath
		res.OriginalName = file.Filename // 使用新上传的原始文件名
		res.FileType = existingFile.FileType
		res.MimeType = existingFile.MimeType
		res.Md5 = md5Value

		// 更新原始文件名（可选，如果需要记录最新的原始名称）
		if existingFile.OriginalName != file.Filename {
			_, err = dao.File.Ctx(ctx).Data(g.Map{
				"original_name": file.Filename,
			}).Where("id", existingFile.Id).Update()
			if err != nil {
				g.Log().Warningf(ctx, "更新原始文件名失败: %v", err)
			}
		}

		// 添加一个消息，告知前端文件已存在，使用自定义状态码
		return res, gerror.New("文件已存在，无需重复上传")
	}

	// 文件不存在，继续保存新文件
	// 生成存储路径
	uploadPath := "uploads/files/" + fileType
	if !gfile.Exists(uploadPath) {
		err = gfile.Mkdir(uploadPath)
		if err != nil {
			return nil, gerror.Newf("创建上传目录失败: %v", err)
		}
	}

	// 生成唯一文件名
	now := gtime.Now()
	uniqueName := now.Format("YmdHis") + "_" + grand.S(8) + fileExt
	filePath := gfile.Join(uploadPath, uniqueName)

	// 重新打开源文件
	src, err = file.Open()
	if err != nil {
		return nil, gerror.Newf("重新打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, gerror.Newf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	written, err = io.Copy(dst, src)
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

	// 创建文件记录
	result, err := dao.File.Ctx(ctx).Data(g.Map{
		"name":          uniqueName,
		"original_name": file.Filename,
		"file_type":     fileType,
		"file_size":     written,
		"file_path":     filePath,
		"mime_type":     mimeType,
		"md5":           md5Value,
		"status":        1,
	}).Insert()

	if err != nil {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.Newf("保存文件信息失败: %v", err)
	}

	// 获取插入ID
	fileId, err := result.LastInsertId()
	if err != nil {
		// 删除已上传的文件
		_ = gfile.Remove(filePath)
		return nil, gerror.Newf("获取文件ID失败: %v", err)
	}

	res.FileId = fileId
	res.FileName = uniqueName
	res.FileSize = written
	res.FilePath = filePath
	res.OriginalName = file.Filename
	res.FileType = fileType
	res.MimeType = mimeType
	res.Md5 = md5Value

	g.Log().Infof(ctx, "文件上传成功，新ID: %d, MD5: %s", fileId, md5Value)
	return res, nil
}

// BatchPushByDevices 批量推送文件到设备
func (s *fileService) BatchPushByDevices(ctx context.Context, req *v1.BatchPushByDevicesReq) (res *v1.BatchPushByDevicesRes, err error) {
	res = &v1.BatchPushByDevicesRes{}

	// 验证并调整并发数
	if req.MaxWorker <= 0 {
		req.MaxWorker = 1
	} else if req.MaxWorker > 50 {
		req.MaxWorker = 50
	}

	// 查询文件信息
	var fileInfo *model.File
	err = dao.File.Ctx(ctx).Where("id", req.FileId).Scan(&fileInfo)
	if err != nil {
		return nil, err
	}
	if fileInfo == nil {
		return nil, gerror.New("文件不存在")
	}

	// 检查文件是否存在
	if !gfile.Exists(fileInfo.FilePath) {
		return nil, gerror.New("文件不存在或已被删除")
	}

	// 检查设备列表是否为空
	if len(req.DeviceIds) == 0 {
		return nil, gerror.New("设备列表不能为空")
	}

	// 生成任务ID
	taskId := fmt.Sprintf("file_push_%s", grand.S(16))

	// 创建任务
	task := &FileBatchTask{
		TaskId:    taskId,
		Status:    "pending",
		Total:     len(req.DeviceIds),
		Completed: 0,
		Failed:    0,
		Results:   make([]v1.BatchTaskResult, 0),
	}

	// 注册任务
	fileTaskManager.mutex.Lock()
	fileTaskManager.tasks[taskId] = task
	fileTaskManager.mutex.Unlock()

	// 启动后台任务
	go s.doBatchPushTask(ctx, task, fileInfo, req.DeviceIds, req.MaxWorker)

	// 构建响应
	res.TaskId = taskId
	res.Total = len(req.DeviceIds)
	res.DeviceIds = req.DeviceIds

	return
}

// BatchTaskStatus 查询批量任务状态
func (s *fileService) BatchTaskStatus(ctx context.Context, req *v1.BatchTaskStatusReq) (res *v1.BatchTaskStatusRes, err error) {
	// 查找任务
	fileTaskManager.mutex.Lock()
	task, exists := fileTaskManager.tasks[req.TaskId]
	fileTaskManager.mutex.Unlock()

	if !exists {
		return nil, gerror.Newf("任务 %s 不存在", req.TaskId)
	}

	// 构建响应
	res = &v1.BatchTaskStatusRes{
		TaskId:    task.TaskId,
		Status:    task.Status,
		Total:     task.Total,
		Completed: task.Completed,
		Failed:    task.Failed,
		Results:   task.Results,
	}

	// 清理已完成超过30分钟的任务
	go s.cleanupCompletedTasks()

	return
}

// doBatchPushTask 执行批量推送任务
func (s *fileService) doBatchPushTask(ctx context.Context, task *FileBatchTask, fileInfo *model.File, deviceIds []string, maxWorker int) {
	// 更新任务状态
	task.Status = "running"

	// 创建同步通道
	ch := make(chan struct{}, maxWorker)
	resultCh := make(chan struct {
		deviceId string
		err      error
	}, len(deviceIds))

	// 获取ADB实例
	adbInstance := adb.Default()

	// 设备上的目标路径 - 注意：确保使用Unix风格的路径分隔符
	devicePath := DeviceTargetDir + "/" + fileInfo.OriginalName
	// g.Log().Debugf(ctx, "文件推送目标路径: %s", devicePath)

	// 启动goroutine执行操作
	for _, deviceId := range deviceIds {
		ch <- struct{}{} // 获取信号量
		go func(deviceId string) {
			defer func() {
				<-ch // 释放信号量
			}()

			// 调用ADB命令推送文件
			err := adbInstance.PushFile(deviceId, fileInfo.FilePath, devicePath)
			resultCh <- struct {
				deviceId string
				err      error
			}{deviceId: deviceId, err: err}
		}(deviceId)
	}

	// 收集结果
	for i := 0; i < len(deviceIds); i++ {
		result := <-resultCh
		task.Mutex.Lock()
		if result.err != nil {
			task.Failed++
			task.Results = append(task.Results, v1.BatchTaskResult{
				DeviceId: result.deviceId,
				Status:   "failed",
				Message:  result.err.Error(),
			})
			g.Log().Errorf(ctx, "推送文件到设备 %s 失败: %v", result.deviceId, result.err)
		} else {
			task.Completed++
			task.Results = append(task.Results, v1.BatchTaskResult{
				DeviceId: result.deviceId,
				Status:   "complete",
				Message:  fmt.Sprintf("文件已成功推送到 %s", devicePath),
			})
			g.Log().Infof(ctx, "推送文件到设备 %s 成功", result.deviceId)
		}
		task.Mutex.Unlock()
	}

	// 更新任务状态
	task.Status = "complete"
}

// cleanupCompletedTasks 清理已完成的任务
func (s *fileService) cleanupCompletedTasks() {
	fileTaskManager.mutex.Lock()
	defer fileTaskManager.mutex.Unlock()

	for id, task := range fileTaskManager.tasks {
		if (task.Status == "complete" || task.Status == "failed") && len(task.Results) > 0 {
			// 检查第一个结果的时间戳，如果超过30分钟，则删除任务
			// 由于没有直接存储时间，这里简化处理，只保留最近的100个任务
			if len(fileTaskManager.tasks) > 100 {
				delete(fileTaskManager.tasks, id)
			}
		}
	}
}

// getFileTypeByExt 根据扩展名获取文件类型
func (s *fileService) getFileTypeByExt(ext string) string {
	ext = strings.ToLower(ext)

	// 图片类型
	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true}
	// 文档类型
	docExts := map[string]bool{".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".pdf": true, ".txt": true}
	// 视频类型
	videoExts := map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".mkv": true}
	// 音频类型
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".aac": true}
	// 压缩包类型
	archiveExts := map[string]bool{".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true}
	// 应用类型
	appExts := map[string]bool{".apk": true, ".ipa": true, ".exe": true, ".dmg": true}

	switch {
	case imageExts[ext]:
		return "image"
	case docExts[ext]:
		return "document"
	case videoExts[ext]:
		return "video"
	case audioExts[ext]:
		return "audio"
	case archiveExts[ext]:
		return "archive"
	case appExts[ext]:
		return "app"
	default:
		return "other"
	}
}

// List 获取文件列表
func (s *fileService) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	res = &v1.ListRes{
		List: make([]v1.File, 0),
	}

	// 查询数据库中的文件列表
	fileModel := dao.File.Ctx(ctx)

	// 如果指定了原始文件名，添加查询条件
	if req.OriginalName != "" {
		fileModel = fileModel.WhereLike(dao.File.Columns().OriginalName, "%"+req.OriginalName+"%")
	}

	var files []*model.File
	err = fileModel.Page(req.Page, req.PageSize).Order("id DESC").Scan(&files)
	if err != nil {
		return nil, err
	}

	// 获取总数
	countModel := dao.File.Ctx(ctx)
	// 同样应用原始文件名过滤条件
	if req.OriginalName != "" {
		countModel = countModel.WhereLike(dao.File.Columns().OriginalName, "%"+req.OriginalName+"%")
	}

	count, err := countModel.Count()
	if err != nil {
		return nil, err
	}
	res.Total = int(count)

	// 转换为API返回结构
	for _, file := range files {
		res.List = append(res.List, v1.File{
			FileId:       file.Id,
			FileName:     file.Name,
			OriginalName: file.OriginalName,
			FileType:     file.FileType,
			FileSize:     file.FileSize,
			UpdatedAt:    file.UpdatedAt.String(),
		})
	}

	return
}

// Create 创建文件记录
func (s *fileService) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	res = &v1.CreateRes{}

	// 创建文件记录
	result, err := dao.File.Ctx(ctx).Data(g.Map{
		"name":      req.FileName,
		"file_size": req.FileSize,
		"file_type": "unknown", // 可以根据需要设置
		"status":    1,         // 默认状态为正常
	}).Insert()

	if err != nil {
		return nil, gerror.Newf("创建文件记录失败: %v", err)
	}

	// 获取插入ID
	fileId, err := result.LastInsertId()
	if err != nil {
		return nil, gerror.Newf("获取文件ID失败: %v", err)
	}

	res.FileId = fileId
	return
}

// Delete 删除文件
func (s *fileService) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	res = &v1.DeleteRes{}

	// 查询文件信息
	var fileInfo *model.File
	err = dao.File.Ctx(ctx).Where("id", req.FileId).Scan(&fileInfo)
	if err != nil {
		return nil, err
	}
	if fileInfo == nil {
		return nil, gerror.New("文件不存在")
	}

	// 如果文件存在，则删除物理文件
	if fileInfo.FilePath != "" && gfile.Exists(fileInfo.FilePath) {
		err = gfile.Remove(fileInfo.FilePath)
		if err != nil {
			return nil, gerror.Newf("删除文件失败: %v", err)
		}
	}

	// 从数据库中删除记录
	_, err = dao.File.Ctx(ctx).Where("id", req.FileId).Delete()
	if err != nil {
		return nil, gerror.Newf("删除文件记录失败: %v", err)
	}

	return
}
