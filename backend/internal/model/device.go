package model

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Device 设备信息
type Device struct {
	Id        int64       `json:"id"         description:"设备ID"`
	Name      string      `json:"name"       description:"设备名称"`
	DeviceId  string      `json:"deviceId"   description:"设备唯一标识"`
	Status    string      `json:"status"     description:"设备状态(online/offline)"`
	GroupId   int64       `json:"groupId"    description:"所属分组ID"`
	CreatedAt *gtime.Time `json:"createdAt"  description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt"  description:"更新时间"`
}

// DeviceListReq 设备列表请求参数
type DeviceListReq struct {
	Page     int    `json:"page"     v:"required#请输入页码"`   // 页码
	PageSize int    `json:"pageSize" v:"required#请输入每页数量"` // 每页数量
	GroupId  int64  `json:"groupId"`                       // 分组ID
	Keyword  string `json:"keyword"`                       // 搜索关键词
}

// DeviceListRes 设备列表返回结果
type DeviceListRes struct {
	List     []Device `json:"list"`     // 设备列表
	Total    int      `json:"total"`    // 总数
	Page     int      `json:"page"`     // 页码
	PageSize int      `json:"pageSize"` // 每页数量
}
