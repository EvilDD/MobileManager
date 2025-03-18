package service

import (
	"backend/internal/model"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var GroupService = &groupService{}

type groupService struct{}

// List 获取分组列表
func (s *groupService) List(ctx context.Context, req *model.GroupListReq) (*model.GroupListRes, error) {
	m := g.DB().Model("group")

	// 关键词搜索
	if req.Keyword != "" {
		m = m.Where("name LIKE ?", "%"+req.Keyword+"%")
	}

	// 获取总数
	count, err := m.Count()
	if err != nil {
		return nil, err
	}

	// 获取列表
	var list []model.Group
	err = m.Page(req.Page, req.PageSize).Order("id DESC").Scan(&list)
	if err != nil {
		return nil, err
	}

	return &model.GroupListRes{
		List:     list,
		Total:    count,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Create 创建分组
func (s *groupService) Create(ctx context.Context, req *model.GroupCreateReq) error {
	_, err := g.DB().Model("group").Data(g.Map{
		"name":        req.Name,
		"description": req.Description,
	}).Insert()
	return err
}

// Update 更新分组
func (s *groupService) Update(ctx context.Context, req *model.GroupUpdateReq) error {
	_, err := g.DB().Model("group").Data(g.Map{
		"name":        req.Name,
		"description": req.Description,
	}).Where("id", req.Id).Update()
	return err
}

// Delete 删除分组
func (s *groupService) Delete(ctx context.Context, id int64) error {
	// 检查分组下是否有设备
	count, err := g.DB().Model("device").Where("group_id", id).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("该分组下存在设备，无法删除")
	}

	_, err = g.DB().Model("group").Where("id", id).Delete()
	return err
}
