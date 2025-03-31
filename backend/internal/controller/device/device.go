package device

import (
	"context"

	v1 "backend/api/device/v1"
	"backend/internal/service"
)

// List 获取设备列表
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	return service.DeviceService.List(ctx, req)
}

// Create 创建设备
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	if err = service.DeviceService.Create(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CreateRes{}, nil
}

// Update 更新设备
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	if err = service.DeviceService.Update(ctx, req); err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}

// Delete 删除设备
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = service.DeviceService.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}

// BatchGoHome 批量回到主菜单
func (c *ControllerV1) BatchGoHome(ctx context.Context, req *v1.BatchGoHomeReq) (res *v1.BatchGoHomeRes, err error) {
	results, err := service.DeviceService.BatchGoHome(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.BatchGoHomeRes{Results: results}, nil
}

// BatchKillApps 批量清除当前应用
func (c *ControllerV1) BatchKillApps(ctx context.Context, req *v1.BatchKillAppsReq) (res *v1.BatchKillAppsRes, err error) {
	results, err := service.DeviceService.BatchKillApps(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.BatchKillAppsRes{Results: results}, nil
}
