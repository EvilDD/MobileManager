package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceStatus 设备状态枚举
const (
	DeviceStatusOnline  = "online"
	DeviceStatusOffline = "offline"
)

// 分组选项，用于前端下拉选择
type GroupOption struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type ListReq struct {
	g.Meta   `path:"/devices/list" tags:"设备管理" method:"get" summary:"获取设备列表"`
	Page     int    `json:"page" v:"required#请输入页码" dc:"页码"`
	PageSize int    `json:"pageSize" v:"required#请输入每页数量" dc:"每页数量"`
	GroupId  int64  `json:"groupId" dc:"分组ID"`
	Keyword  string `json:"keyword" dc:"搜索关键词"`
}

type ListRes struct {
	List         []Device      `json:"list" dc:"设备列表"`
	Page         int           `json:"page" dc:"页码"`
	PageSize     int           `json:"pageSize" dc:"每页数量"`
	Total        int           `json:"total" dc:"总数"`
	GroupOptions []GroupOption `json:"groupOptions" dc:"分组选项列表"`
}

type CreateReq struct {
	g.Meta   `path:"/devices/create" tags:"设备管理" method:"post" summary:"创建设备"`
	Name     string `json:"name" v:"required#请输入设备名称" dc:"设备名称"`
	DeviceId string `json:"deviceId" v:"required|regex:^[a-zA-Z0-9_\\-\\:\\.]+$#请输入设备ID|设备ID只能包含字母、数字、下划线、中划线、冒号和点" dc:"设备ID，只能包含字母、数字、下划线、中划线、冒号和点"`
	GroupId  int64  `json:"groupId" dc:"分组ID，不指定则为0表示未分组"`
	Status   string `json:"status" v:"required|in:online,offline#请输入设备状态|设备状态只能是online或offline" dc:"设备状态：online-在线，offline-离线"`
}

type CreateRes struct{}

type UpdateReq struct {
	g.Meta   `path:"/devices/update" tags:"设备管理" method:"put" summary:"更新设备"`
	Id       int64  `json:"id" v:"required#请输入设备ID" dc:"设备ID"`
	Name     string `json:"name" v:"required#请输入设备名称" dc:"设备名称"`
	DeviceId string `json:"deviceId" v:"required|regex:^[a-zA-Z0-9_\\-\\:\\.]+$#请输入设备ID|设备ID只能包含字母、数字、下划线、中划线、冒号和点" dc:"设备ID，只能包含字母、数字、下划线、中划线、冒号和点"`
	GroupId  int64  `json:"groupId" dc:"分组ID，不指定则为0表示未分组"`
	Status   string `json:"status" v:"required|in:online,offline#请输入设备状态|设备状态只能是online或offline" dc:"设备状态：online-在线，offline-离线"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/devices/delete" tags:"设备管理" method:"delete" summary:"删除设备"`
	Id     int64 `json:"id" v:"required#请输入设备ID" dc:"设备ID"`
}

type DeleteRes struct{}

type Device struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	DeviceId  string `json:"deviceId"`
	GroupId   int64  `json:"groupId"`
	GroupName string `json:"groupName"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// BatchGoHomeReq 批量回到主菜单请求
type BatchGoHomeReq struct {
	g.Meta    `path:"/devices/batch-go-home" tags:"设备管理" method:"post" summary:"批量回到主菜单"`
	DeviceIds []string `json:"deviceIds" v:"required#请选择设备|length:1,50#设备数量必须在1-50之间" dc:"设备ID列表，最多50个"`
}

type BatchGoHomeRes struct {
	Results map[string]string `json:"results" dc:"操作结果，key为设备ID，value为错误信息（成功为空）"`
}

// BatchKillAppsReq 批量清除当前应用请求
type BatchKillAppsReq struct {
	g.Meta    `path:"/devices/batch-kill-apps" tags:"设备管理" method:"post" summary:"批量清除当前应用"`
	DeviceIds []string `json:"deviceIds" v:"required#请选择设备|length:1,50#设备数量必须在1-50之间" dc:"设备ID列表，最多50个"`
}

type BatchKillAppsRes struct {
	Results map[string]string `json:"results" dc:"操作结果，key为设备ID，value为错误信息（成功为空）"`
}
