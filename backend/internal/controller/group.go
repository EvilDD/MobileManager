package controller

import (
	"backend/internal/model"
	"backend/internal/service"
	"context"
)

var GroupController = &groupController{}

type groupController struct{}

// List 获取分组列表
func (c *groupController) List(ctx context.Context, req *model.GroupListReq) (res *model.GroupListRes, err error) {
	return service.GroupService.List(ctx, req)
}

// Create 创建分组
func (c *groupController) Create(ctx context.Context, req *model.GroupCreateReq) (res *struct{}, err error) {
	return &struct{}{}, service.GroupService.Create(ctx, req)
}

// Update 更新分组
func (c *groupController) Update(ctx context.Context, req *model.GroupUpdateReq) (res *struct{}, err error) {
	return &struct{}{}, service.GroupService.Update(ctx, req)
}

// Delete 删除分组
func (c *groupController) Delete(ctx context.Context, req *struct {
	Id int64 `v:"required#请输入分组ID"`
}) (res *struct{}, err error) {
	return &struct{}{}, service.GroupService.Delete(ctx, req.Id)
}
