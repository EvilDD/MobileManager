package model

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Group 分组信息
type Group struct {
	Id          int64       `json:"id"          description:"分组ID"`
	Name        string      `json:"name"        description:"分组名称"`
	Description string      `json:"description" description:"分组描述"`
	CreatedAt   *gtime.Time `json:"createdAt"   description:"创建时间"`
	UpdatedAt   *gtime.Time `json:"updatedAt"   description:"更新时间"`
}

// GroupListReq 分组列表请求参数
type GroupListReq struct {
	Page     int    `json:"page"     v:"required#请输入页码"`   // 页码
	PageSize int    `json:"pageSize" v:"required#请输入每页数量"` // 每页数量
	Keyword  string `json:"keyword"`                       // 搜索关键词
}

// GroupListRes 分组列表返回结果
type GroupListRes struct {
	List     []Group `json:"list"`     // 分组列表
	Total    int     `json:"total"`    // 总数
	Page     int     `json:"page"`     // 页码
	PageSize int     `json:"pageSize"` // 每页数量
}

// GroupCreateReq 创建分组请求参数
type GroupCreateReq struct {
	Name        string `json:"name"        v:"required#请输入分组名称"` // 分组名称
	Description string `json:"description"`                      // 分组描述
}

// GroupUpdateReq 更新分组请求参数
type GroupUpdateReq struct {
	Id          int64  `json:"id"          v:"required#请输入分组ID"` // 分组ID
	Name        string `json:"name"        v:"required#请输入分组名称"` // 分组名称
	Description string `json:"description"`                      // 分组描述
}
