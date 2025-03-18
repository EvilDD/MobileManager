package controller

import (
	"backend/internal/model"
	"backend/internal/service"
	"context"
)

var DeviceController = &deviceController{}

type deviceController struct{}

// List 获取设备列表
func (c *deviceController) List(ctx context.Context, req *model.DeviceListReq) (res *model.DeviceListRes, err error) {
	return service.DeviceService.List(ctx, req)
}
