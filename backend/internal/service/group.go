package service

import (
	"context"

	v1 "backend/api/group/v1"
	"backend/internal/dao"
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var GroupService = groupService{}

type groupService struct{}

// List 获取分组列表
func (s *groupService) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	res = &v1.ListRes{
		List:     make([]v1.Group, 0),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	m := dao.Group.Ctx(ctx)
	if req.Keyword != "" {
		m = m.WhereLike("name", "%"+req.Keyword+"%")
	}

	res.Total, err = m.Count()
	if err != nil {
		return nil, err
	}

	var groups []*entity.Group
	err = m.Page(req.Page, req.PageSize).Scan(&groups)
	if err != nil {
		return nil, err
	}

	// group.Id 为 0 时，表示新设备
	deviceCount, err := dao.Device.Ctx(ctx).Where("group_id", 0).Count()
	if err != nil {
		return nil, err
	}

	res.List = append(res.List, v1.Group{
		Id:          0,
		Name:        "新设备",
		Description: "未分组的设备,请及时分组",
		DeviceCount: deviceCount,
		CreatedAt:   "",
		UpdatedAt:   "",
	})

	// 获取每个分组的设备数量
	for _, group := range groups {
		// 查询设备数量
		deviceCount, err := dao.Device.Ctx(ctx).Where("group_id", group.Id).Count()
		if err != nil {
			return nil, err
		}

		res.List = append(res.List, v1.Group{
			Id:          int64(group.Id),
			Name:        group.Name,
			Description: group.Description,
			DeviceCount: deviceCount,
			CreatedAt:   group.CreatedAt.String(),
			UpdatedAt:   group.UpdatedAt.String(),
		})
	}

	return
}

// Create 创建分组
func (s *groupService) Create(ctx context.Context, req *v1.CreateReq) error {
	_, err := dao.Group.Ctx(ctx).Fields("name", "description").Data(g.Map{
		"name":        req.Name,
		"description": req.Description,
	}).Insert()
	return err
}

// Update 更新分组
func (s *groupService) Update(ctx context.Context, req *v1.UpdateReq) error {
	_, err := dao.Group.Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"name":        req.Name,
		"description": req.Description,
	}).Update()
	return err
}

// Delete 删除分组
func (s *groupService) Delete(ctx context.Context, req *v1.DeleteReq) error {
	// 检查分组下是否有设备
	count, err := dao.Device.Ctx(ctx).Where("group_id", req.Id).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err = dao.Group.Ctx(ctx).Where("id", req.Id).Delete()
	return err
}

// BatchUpdateDevicesGroup 批量修改设备分组
func (s *groupService) BatchUpdateDevicesGroup(ctx context.Context, req *v1.BatchUpdateDevicesGroupReq) error {
	// 检查分组是否存在
	count, err := dao.Group.Ctx(ctx).Where("id", req.GroupId).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("分组不存在")
	}

	// 批量更新设备的分组ID
	_, err = dao.Device.Ctx(ctx).
		WhereIn("id", req.DeviceIds).
		Data(g.Map{
			"group_id": req.GroupId,
		}).
		Update()
	return err
}
