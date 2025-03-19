package service

import (
	"context"

	v1 "backend/api/device/v1"
	"backend/internal/dao"
	"backend/internal/model"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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
	if req.GroupId >= 0 {
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

// Create 创建设备
func (s *deviceService) Create(ctx context.Context, req *v1.CreateReq) error {
	// 如果指定了分组ID，则检查分组是否存在
	if req.GroupId > 0 {
		count, err := dao.Group.Ctx(ctx).Where("id", req.GroupId).Count()
		if err != nil {
			return err
		}
		if count == 0 {
			return gerror.New("分组不存在")
		}
	}

	// 检查设备ID是否已存在
	count, err := dao.Device.Ctx(ctx).Where("device_id", req.DeviceId).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("设备ID已存在")
	}

	// 如果未指定状态，则默认为离线状态
	status := req.Status
	if status == "" {
		status = v1.DeviceStatusOffline
	}

	_, err = dao.Device.Ctx(ctx).Data(g.Map{
		"name":      req.Name,
		"device_id": req.DeviceId,
		"group_id":  req.GroupId,
		"status":    status,
	}).Insert()
	return err
}

// Update 更新设备
func (s *deviceService) Update(ctx context.Context, req *v1.UpdateReq) error {
	// 如果指定了分组ID，则检查分组是否存在
	if req.GroupId > 0 {
		count, err := dao.Group.Ctx(ctx).Where("id", req.GroupId).Count()
		if err != nil {
			return err
		}
		if count == 0 {
			return gerror.New("分组不存在")
		}
	}

	// 检查设备ID是否与其他设备重复
	count, err := dao.Device.Ctx(ctx).Where("device_id", req.DeviceId).WhereNot("id", req.Id).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("设备ID已被其他设备使用")
	}

	// 检查设备是否存在
	count, err = dao.Device.Ctx(ctx).Where("id", req.Id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("设备不存在")
	}

	// 如果未指定状态，则默认为离线状态
	status := req.Status
	if status == "" {
		status = v1.DeviceStatusOffline
	}

	_, err = dao.Device.Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"name":      req.Name,
		"device_id": req.DeviceId,
		"group_id":  req.GroupId,
		"status":    status,
	}).Update()
	return err
}

// Delete 删除设备
func (s *deviceService) Delete(ctx context.Context, req *v1.DeleteReq) error {
	// 检查设备是否存在
	count, err := dao.Device.Ctx(ctx).Where("id", req.Id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("设备不存在")
	}

	_, err = dao.Device.Ctx(ctx).Where("id", req.Id).Delete()
	return err
}
