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
