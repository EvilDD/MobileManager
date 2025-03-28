package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta   `path:"/groups/list" tags:"分组管理" method:"get" summary:"获取分组列表"`
	Page     int    `json:"page" v:"required#请输入页码" dc:"页码"`
	PageSize int    `json:"pageSize" v:"required#请输入每页数量" dc:"每页数量"`
	Keyword  string `json:"keyword" dc:"搜索关键词"`
}

type ListRes struct {
	List     []Group `json:"list" dc:"分组列表"`
	Page     int     `json:"page" dc:"页码"`
	PageSize int     `json:"pageSize" dc:"每页数量"`
	Total    int     `json:"total" dc:"总数"`
}

type CreateReq struct {
	g.Meta      `path:"/groups/create" tags:"分组管理" method:"post" summary:"创建分组"`
	Name        string `json:"name" v:"required#请输入分组名称" dc:"分组名称"`
	Description string `json:"description" dc:"分组描述"`
}

type CreateRes struct{}

type UpdateReq struct {
	g.Meta      `path:"/groups/update" tags:"分组管理" method:"put" summary:"更新分组"`
	Id          int64  `json:"id" v:"required#请输入分组ID" dc:"分组ID"`
	Name        string `json:"name" v:"required#请输入分组名称" dc:"分组名称"`
	Description string `json:"description" dc:"分组描述"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/groups/delete" tags:"分组管理" method:"delete" summary:"删除分组"`
	Id     int64 `json:"id" v:"required#请输入分组ID" dc:"分组ID"`
}

type DeleteRes struct{}

type BatchUpdateDevicesGroupReq struct {
	g.Meta    `path:"/groups/batch-update-devices" tags:"分组管理" method:"put" summary:"批量修改设备分组"`
	GroupId   int64   `json:"groupId" v:"required#请选择目标分组" dc:"目标分组ID"`
	DeviceIds []int64 `json:"deviceIds" v:"required#请选择要修改的设备" dc:"设备ID列表"`
}

type BatchUpdateDevicesGroupRes struct{}

type Group struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DeviceCount int    `json:"deviceCount" dc:"设备数量"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
