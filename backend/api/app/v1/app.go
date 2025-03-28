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
