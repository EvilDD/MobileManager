package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// 应用类型枚举
const (
	AppTypeSystem   = "系统应用"
	AppTypeUser     = "用户应用"
	AppTypeSettings = "系统设置"
)

// 批量操作任务状态
const (
	TaskStatusPending  = "pending"  // 等待执行
	TaskStatusRunning  = "running"  // 执行中
	TaskStatusComplete = "complete" // 执行完成
	TaskStatusFailed   = "failed"   // 执行失败
)

type ListReq struct {
	g.Meta   `path:"/apps/list" tags:"应用管理" method:"get" summary:"获取应用列表"`
	Page     int    `json:"page" v:"required#请输入页码" dc:"页码"`
	PageSize int    `json:"pageSize" v:"required#请输入每页数量" dc:"每页数量"`
	AppType  string `json:"appType" dc:"应用类型"`
	Keyword  string `json:"keyword" dc:"搜索关键词"`
}

type ListRes struct {
	List     []App `json:"list" dc:"应用列表"`
	Page     int   `json:"page" dc:"页码"`
	PageSize int   `json:"pageSize" dc:"每页数量"`
	Total    int   `json:"total" dc:"总数"`
}

type CreateReq struct {
	g.Meta      `path:"/apps/create" tags:"应用管理" method:"post" summary:"创建应用"`
	Name        string `json:"name" v:"required#请输入应用名称" dc:"应用名称"`
	PackageName string `json:"packageName" v:"required#请输入应用包名" dc:"应用包名"`
	Version     string `json:"version" v:"required#请输入版本号" dc:"版本号"`
	Size        int64  `json:"size" v:"required#请输入应用大小" dc:"应用大小(KB)"`
	AppType     string `json:"appType" v:"required#请选择应用类型" dc:"应用类型"`
	ApkPath     string `json:"apkPath" v:"required#请上传APK文件" dc:"APK文件路径"`
}

type CreateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/apps/delete" tags:"应用管理" method:"delete" summary:"删除应用"`
	Id     int64 `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
}

type DeleteRes struct{}

type InstallReq struct {
	g.Meta   `path:"/apps/install" tags:"应用管理" method:"post" summary:"安装应用"`
	Id       int64  `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	DeviceId string `json:"deviceId" v:"required#请输入设备ID" dc:"设备ID"`
}

type InstallRes struct{}

type UninstallReq struct {
	g.Meta   `path:"/apps/uninstall" tags:"应用管理" method:"post" summary:"卸载应用"`
	Id       int64  `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	DeviceId string `json:"deviceId" v:"required#请输入设备ID" dc:"设备ID"`
}

type UninstallRes struct{}

type StartReq struct {
	g.Meta   `path:"/apps/start" tags:"应用管理" method:"post" summary:"启动应用"`
	Id       int64  `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	DeviceId string `json:"deviceId" v:"required#请输入设备ID" dc:"设备ID"`
}

type StartRes struct{}

type UploadReq struct {
	g.Meta `path:"/apps/upload" tags:"应用管理" method:"post" summary:"上传APK文件"`
	File   *ghttp.UploadFile `type:"file" json:"file" v:"required#请选择APK文件" dc:"APK文件"`
}

type UploadRes struct {
	FileName string `json:"fileName" dc:"文件名"`
	FileSize int64  `json:"fileSize" dc:"文件大小(KB)"`
	FilePath string `json:"filePath" dc:"文件路径"`
}

type App struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	PackageName string `json:"packageName"`
	Version     string `json:"version"`
	Size        int64  `json:"size"`
	AppType     string `json:"appType"`
	ApkPath     string `json:"apkPath"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// 批量安装请求
type BatchInstallReq struct {
	g.Meta    `path:"/apps/batch-install" tags:"应用管理" method:"post" summary:"批量安装应用" description:"按分组批量安装应用到设备，支持并发控制"`
	Id        int64 `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	GroupId   int64 `json:"groupId" v:"required#请输入分组ID" dc:"分组ID"`
	MaxWorker int   `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

type BatchInstallRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}

// 批量卸载请求
type BatchUninstallReq struct {
	g.Meta    `path:"/apps/batch-uninstall" tags:"应用管理" method:"post" summary:"批量卸载应用" description:"按分组批量卸载设备上的应用，支持并发控制"`
	Id        int64 `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	GroupId   int64 `json:"groupId" v:"required#请输入分组ID" dc:"分组ID"`
	MaxWorker int   `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

type BatchUninstallRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}

// 批量启动请求
type BatchStartReq struct {
	g.Meta    `path:"/apps/batch-start" tags:"应用管理" method:"post" summary:"批量启动应用" description:"按分组批量启动设备上的应用，支持并发控制"`
	Id        int64 `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	GroupId   int64 `json:"groupId" v:"required#请输入分组ID" dc:"分组ID"`
	MaxWorker int   `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

type BatchStartRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}

// 批量操作任务查询请求
type BatchTaskStatusReq struct {
	g.Meta `path:"/apps/batch-task-status" tags:"应用管理" method:"get" summary:"查询批量操作任务状态" description:"查询批量安装/卸载/启动任务的执行状态和结果"`
	TaskId string `v:"required#请输入任务ID" in:"query" dc:"任务ID"`
}

type BatchTaskStatusRes struct {
	TaskId    string            `json:"taskId" dc:"任务ID"`
	Status    string            `json:"status" dc:"任务状态(pending:等待执行 running:执行中 complete:执行完成 failed:执行失败)"`
	Total     int               `json:"total" dc:"总设备数"`
	Completed int               `json:"completed" dc:"已完成数"`
	Failed    int               `json:"failed" dc:"失败数"`
	Results   []BatchTaskResult `json:"results" dc:"执行结果列表"`
}

type BatchTaskResult struct {
	DeviceId string `json:"deviceId" dc:"设备ID"`
	Status   string `json:"status" dc:"执行状态(complete:成功 failed:失败)"`
	Message  string `json:"message" dc:"执行结果信息"`
}

// 按设备ID批量安装请求
type BatchInstallByDevicesReq struct {
	g.Meta    `path:"/apps/batch-install-by-devices" tags:"应用管理" method:"post" summary:"按设备批量安装应用" description:"按设备ID列表批量安装应用，支持并发控制"`
	Id        int64    `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	DeviceIds []string `json:"deviceIds" v:"required|length:1,50#请选择设备|设备数量必须在1-50之间" dc:"设备ID列表，最多50个"`
	MaxWorker int      `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

type BatchInstallByDevicesRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}

// 按设备ID批量卸载请求
type BatchUninstallByDevicesReq struct {
	g.Meta    `path:"/apps/batch-uninstall-by-devices" tags:"应用管理" method:"post" summary:"按设备批量卸载应用" description:"按设备ID列表批量卸载应用，支持并发控制"`
	Id        int64    `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	DeviceIds []string `json:"deviceIds" v:"required|length:1,50#请选择设备|设备数量必须在1-50之间" dc:"设备ID列表，最多50个"`
	MaxWorker int      `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

type BatchUninstallByDevicesRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}

// 按设备ID批量启动请求
type BatchStartByDevicesReq struct {
	g.Meta    `path:"/apps/batch-start-by-devices" tags:"应用管理" method:"post" summary:"按设备批量启动应用" description:"按设备ID列表批量启动应用，支持并发控制"`
	Id        int64    `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	DeviceIds []string `json:"deviceIds" v:"required|length:1,50#请选择设备|设备数量必须在1-50之间" dc:"设备ID列表，最多50个"`
	MaxWorker int      `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

type BatchStartByDevicesRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}

// 按设备ID批量停止请求
type BatchStopByDevicesReq struct {
	g.Meta    `path:"/apps/batch-stop-by-devices" tags:"应用管理" method:"post" summary:"按设备批量停止应用" description:"按设备ID列表批量停止应用，支持并发控制"`
	Id        int64    `json:"id" v:"required#请输入应用ID" dc:"应用ID"`
	DeviceIds []string `json:"deviceIds" v:"required|length:1,50#请选择设备|设备数量必须在1-50之间" dc:"设备ID列表，最多50个"`
	MaxWorker int      `json:"maxWorker" v:"required|min:1|max:50#请输入并发数|并发数最小为1|并发数最大为50" dc:"最大并发数(1-50)"`
}

type BatchStopByDevicesRes struct {
	TaskId    string   `json:"taskId" dc:"任务ID(用于查询任务状态)"`
	Total     int      `json:"total" dc:"总设备数"`
	DeviceIds []string `json:"deviceIds" dc:"设备ID列表"`
}
