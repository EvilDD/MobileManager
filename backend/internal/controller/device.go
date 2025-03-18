package controller

import (
	"backend/internal/model"
	"backend/internal/service"
	"context"
)

var DeviceController = &deviceController{}

type deviceController struct{}

// List 获取设备列表
//
// @Summary 获取设备列表
// @Description 获取云手机设备列表，支持分页、关键字搜索和分组筛选
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param groupId query int false "分组ID"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} model.DeviceListRes "返回结果"
// @Router /api/v1/devices/list [GET]
func (c *deviceController) List(ctx context.Context, req *model.DeviceListReq) (res *model.DeviceListRes, err error) {
	return service.DeviceService.List(ctx, req)
}
