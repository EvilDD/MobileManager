package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta   `path:"/devices/list" tags:"设备管理" method:"get" summary:"获取设备列表"`
	Page     int    `json:"page" v:"required#请输入页码" dc:"页码"`
	PageSize int    `json:"pageSize" v:"required#请输入每页数量" dc:"每页数量"`
	GroupId  int64  `json:"groupId" dc:"分组ID"`
	Keyword  string `json:"keyword" dc:"搜索关键词"`
}

type ListRes struct {
	List     []Device `json:"list" dc:"设备列表"`
	Page     int      `json:"page" dc:"页码"`
	PageSize int      `json:"pageSize" dc:"每页数量"`
	Total    int      `json:"total" dc:"总数"`
}

type Device struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	DeviceId  string `json:"deviceId"`
	GroupId   int64  `json:"groupId"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
