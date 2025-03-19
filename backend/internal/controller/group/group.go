package group

import (
	"context"

	v1 "backend/api/group/v1"
	"backend/internal/service"
)

// List 获取分组列表
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	return service.GroupService.List(ctx, req)
}

// Create 创建分组
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	if err = service.GroupService.Create(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CreateRes{}, nil
}

// Update 更新分组
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	if err = service.GroupService.Update(ctx, req); err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}

// Delete 删除分组
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = service.GroupService.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
