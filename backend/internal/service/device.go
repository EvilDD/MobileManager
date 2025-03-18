package service

import (
	"backend/internal/model"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

var DeviceService = &deviceService{}

type deviceService struct{}

// List 获取设备列表
func (s *deviceService) List(ctx context.Context, req *model.DeviceListReq) (*model.DeviceListRes, error) {
	m := g.DB().Model("device")

	// 分组筛选
	if req.GroupId > 0 {
		m = m.Where("group_id", req.GroupId)
	}

	// 关键词搜索
	if req.Keyword != "" {
		m = m.Where("name LIKE ? OR device_id LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 获取总数
	count, err := m.Count()
	if err != nil {
		return nil, err
	}

	// 获取列表
	var list []model.Device
	err = m.Page(req.Page, req.PageSize).Order("id DESC").Scan(&list)
	if err != nil {
		return nil, err
	}

	return &model.DeviceListRes{
		List:     list,
		Total:    count,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
