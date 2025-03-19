package service

import (
	"context"

	v1 "backend/api/device/v1"
	"backend/internal/dao"
	"backend/internal/model"
)

var DeviceService = deviceService{}

type deviceService struct{}

// List 获取设备列表
func (s *deviceService) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	res = &v1.ListRes{
		List:     make([]v1.Device, 0),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	m := dao.Device.Ctx(ctx)
	if req.GroupId > 0 {
		m = m.Where("group_id", req.GroupId)
	}
	if req.Keyword != "" {
		m = m.WhereLike("name", "%"+req.Keyword+"%")
	}

	res.Total, err = m.Count()
	if err != nil {
		return nil, err
	}

	var devices []*model.Device
	err = m.Page(req.Page, req.PageSize).Scan(&devices)
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		res.List = append(res.List, v1.Device{
			Id:        device.Id,
			Name:      device.Name,
			DeviceId:  device.DeviceId,
			GroupId:   device.GroupId,
			Status:    device.Status,
			CreatedAt: device.CreatedAt.String(),
			UpdatedAt: device.UpdatedAt.String(),
		})
	}

	return
}
