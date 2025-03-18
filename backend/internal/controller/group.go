package controller

import (
	"backend/internal/model"
	"backend/internal/service"
	"context"
)

var GroupController = &groupController{}

type groupController struct{}

// List 获取分组列表
//
// @Summary 获取分组列表
// @Description 获取云手机分组列表，支持分页和关键字搜索
// @Tags 分组管理
// @Accept json
// @Produce json
// @Param req body model.GroupListReq true "请求参数"
// @Success 200 {object} model.GroupListRes "返回结果"
// @Router /api/v1/groups/list [GET]
func (c *groupController) List(ctx context.Context, req *model.GroupListReq) (res *model.GroupListRes, err error) {
	return service.GroupService.List(ctx, req)
}

// Create 创建分组
//
// @Summary 创建分组
// @Description 创建新的云手机分组
// @Tags 分组管理
// @Accept json
// @Produce json
// @Param req body model.GroupCreateReq true "请求参数"
// @Success 200 {object} object "返回结果"
// @Router /api/v1/groups/create [POST]
func (c *groupController) Create(ctx context.Context, req *model.GroupCreateReq) (res *struct{}, err error) {
	return &struct{}{}, service.GroupService.Create(ctx, req)
}

// Update 更新分组
//
// @Summary 更新分组
// @Description 更新云手机分组信息
// @Tags 分组管理
// @Accept json
// @Produce json
// @Param req body model.GroupUpdateReq true "请求参数"
// @Success 200 {object} object "返回结果"
// @Router /api/v1/groups/update [PUT]
func (c *groupController) Update(ctx context.Context, req *model.GroupUpdateReq) (res *struct{}, err error) {
	return &struct{}{}, service.GroupService.Update(ctx, req)
}

// GroupDeleteReq 删除分组请求
type GroupDeleteReq struct {
	Id int64 `json:"id" v:"required#请输入分组ID"`
}

// Delete 删除分组
//
// @Summary 删除分组
// @Description 删除云手机分组，只有当分组下没有设备时才能删除
// @Tags 分组管理
// @Accept json
// @Produce json
// @Param req body GroupDeleteReq true "请求参数"
// @Success 200 {object} object "返回结果"
// @Router /api/v1/groups/delete [DELETE]
func (c *groupController) Delete(ctx context.Context, req *GroupDeleteReq) (res *struct{}, err error) {
	return &struct{}{}, service.GroupService.Delete(ctx, req.Id)
}
