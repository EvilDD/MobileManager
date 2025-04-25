package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// 上传文件请求
type UploadReq struct {
	g.Meta `path:"/files/upload" tags:"文件管理" method:"post" summary:"上传文件"`
	File   *ghttp.UploadFile `type:"file" json:"file" v:"required#请选择文件" dc:"文件"`
}

// 上传文件响应
type UploadRes struct {
	FileId       int64  `json:"fileId" dc:"文件ID"`
	FileName     string `json:"fileName" dc:"文件名"`
	FileSize     int64  `json:"fileSize" dc:"文件大小(Byte)"`
	FilePath     string `json:"filePath" dc:"文件路径"`
	OriginalName string `json:"originalName" dc:"原始文件名"`
	FileType     string `json:"fileType" dc:"文件类型"`
	MimeType     string `json:"mimeType" dc:"MIME类型"`
	Md5          string `json:"md5" dc:"MD5值"`
}

// 批量推送文件到设备请求
type BatchPushByDevicesReq struct {
	g.Meta    `path:"/files/batch-push-by-devices" tags:"文件管理" method:"post" summary:"批量推送文件到设备" description:"按设备ID列表批量推送文件，支持并发控制"`
	FileId    int64    `json:"fileId" v:"required#请输入文件ID" dc:"文件ID"`
	DeviceIds []string `json:"deviceIds" v:"required#请选择设备" dc:"设备ID列表，最多50个"`
	MaxWorker int      `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

// 批量推送文件到设备响应
type BatchPushByDevicesRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}

// 批量操作任务状态请求
type BatchTaskStatusReq struct {
	g.Meta `path:"/files/batch-task-status" tags:"文件管理" method:"get" summary:"查询批量操作任务状态" description:"查询批量推送文件任务的执行状态和结果"`
	TaskId string `v:"required#请输入任务ID" in:"query" dc:"任务ID"`
}

// 批量操作任务状态响应
type BatchTaskStatusRes struct {
	TaskId    string            `json:"taskId" dc:"任务ID"`
	Status    string            `json:"status" dc:"任务状态(pending:等待执行 running:执行中 complete:执行完成 failed:执行失败)"`
	Total     int               `json:"total" dc:"总设备数"`
	Completed int               `json:"completed" dc:"已完成数"`
	Failed    int               `json:"failed" dc:"失败数"`
	Results   []BatchTaskResult `json:"results" dc:"执行结果列表"`
}

// 批量任务执行结果
type BatchTaskResult struct {
	DeviceId string `json:"deviceId" dc:"设备ID"`
	Status   string `json:"status" dc:"执行状态(complete:成功 failed:失败)"`
	Message  string `json:"message" dc:"执行结果信息"`
}

// 文件信息
type File struct {
	FileId   int64  `json:"fileId"   v:"required" dc:"文件ID"`
	FileName string `json:"fileName" v:"required" dc:"文件名"`
	FileSize int64  `json:"fileSize" v:"required" dc:"文件大小"`
}

// ListReq 获取文件列表请求
type ListReq struct {
	g.Meta   `path:"/files/list" method:"get" tags:"文件管理" summary:"获取文件列表"`
	Page     int `json:"page" d:"1" v:"min:0#分页号码错误" dc:"分页号码，默认1"`
	PageSize int `json:"pageSize" d:"10" v:"max:50#分页数量最大50条" dc:"分页数量，最大50"`
}

// ListRes 获取文件列表响应
type ListRes struct {
	List  []File `json:"list" dc:"文件列表"`
	Total int    `json:"total" dc:"总数"`
}

// CreateReq 创建文件请求
type CreateReq struct {
	g.Meta   `path:"/files/create" method:"post" tags:"文件管理" summary:"创建文件"`
	FileName string `json:"fileName" v:"required" dc:"文件名"`
	FileSize int64  `json:"fileSize" v:"required" dc:"文件大小"`
}

// CreateRes 创建文件响应
type CreateRes struct {
	FileId int64 `json:"fileId" dc:"文件ID"`
}

// DeleteReq 删除文件请求
type DeleteReq struct {
	g.Meta `path:"/files/delete" method:"delete" tags:"文件管理" summary:"删除文件"`
	FileId int64 `json:"fileId" v:"required" dc:"文件ID"`
}

// DeleteRes 删除文件响应
type DeleteRes struct{}
