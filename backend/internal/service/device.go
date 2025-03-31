package service

import (
	"context"

	v1 "backend/api/device/v1"
	"backend/internal/dao"
	"backend/internal/model"
	"backend/utility/adb"

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

	// 检查请求中是否明确指定了GroupId参数
	if g.RequestFromCtx(ctx).GetQuery("groupId").String() != "" {
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

	// 获取所有分组信息，用于后面关联
	var groups []*model.Group
	err = dao.Group.Ctx(ctx).Scan(&groups)
	if err != nil {
		return nil, err
	}

	// 构建分组ID到分组信息的映射
	groupMap := make(map[int64]model.Group)
	for _, group := range groups {
		groupMap[group.Id] = *group
	}

	// 添加默认的"未分组"选项
	groupMap[0] = model.Group{
		Id:          0,
		Name:        "未分组",
		Description: "未分配分组的设备",
	}

	for _, device := range devices {
		// 查找设备所属分组
		group, exists := groupMap[device.GroupId]
		groupName := "未分组"
		if exists {
			groupName = group.Name
		}

		res.List = append(res.List, v1.Device{
			Id:        device.Id,
			Name:      device.Name,
			DeviceId:  device.DeviceId,
			GroupId:   device.GroupId,
			GroupName: groupName,
			Status:    device.Status,
			CreatedAt: device.CreatedAt.String(),
			UpdatedAt: device.UpdatedAt.String(),
		})
	}

	// 获取所有分组供前端选择使用
	var groupOptions []v1.GroupOption
	for _, group := range groupMap {
		groupOptions = append(groupOptions, v1.GroupOption{
			Id:   group.Id,
			Name: group.Name,
		})
	}
	res.GroupOptions = groupOptions

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
	// 准备更新数据
	updateData := g.Map{
		"name":      req.Name,
		"device_id": req.DeviceId,
		"status":    req.Status,
	}

	// 检查请求中是否明确提供了 groupId 参数
	requestBody := g.RequestFromCtx(ctx).GetBody()
	bodyJson := g.NewVar(requestBody).Map()
	_, hasGroupId := bodyJson["groupId"]

	if hasGroupId {
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
		// 将 group_id 添加到更新数据中
		updateData["group_id"] = req.GroupId
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
	if req.Status == "" {
		updateData["status"] = v1.DeviceStatusOffline
	}

	_, err = dao.Device.Ctx(ctx).Where("id", req.Id).Data(updateData).Update()
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

// BatchGoHome 批量回到主菜单
func (s *deviceService) BatchGoHome(ctx context.Context, req *v1.BatchGoHomeReq) (map[string]string, error) {
	results := make(map[string]string)

	// 并发执行，最多50个并发
	ch := make(chan struct{}, 50)
	resultCh := make(chan struct {
		deviceId string
		err      error
	}, len(req.DeviceIds))

	// 获取ADB实例
	adbInstance := adb.Default()

	// 启动goroutine执行操作
	for _, deviceId := range req.DeviceIds {
		ch <- struct{}{} // 获取信号量
		go func(deviceId string) {
			defer func() {
				<-ch // 释放信号量
			}()

			// 调用ADB命令回到主菜单
			_, err := adbInstance.ExecuteCommand(deviceId, "shell", "input", "keyevent", "3")
			resultCh <- struct {
				deviceId string
				err      error
			}{deviceId: deviceId, err: err}
		}(deviceId)
	}

	// 收集结果
	for i := 0; i < len(req.DeviceIds); i++ {
		result := <-resultCh
		if result.err != nil {
			results[result.deviceId] = result.err.Error()
		} else {
			results[result.deviceId] = ""
		}
	}

	return results, nil
}

// BatchKillApps 批量清除当前应用
func (s *deviceService) BatchKillApps(ctx context.Context, req *v1.BatchKillAppsReq) (map[string]string, error) {
	results := make(map[string]string)

	// 并发执行，最多50个并发
	ch := make(chan struct{}, 50)
	resultCh := make(chan struct {
		deviceId string
		err      error
	}, len(req.DeviceIds))

	// 获取ADB实例
	adbInstance := adb.Default()

	// 启动goroutine执行操作
	for _, deviceId := range req.DeviceIds {
		ch <- struct{}{} // 获取信号量
		go func(deviceId string) {
			defer func() {
				<-ch // 释放信号量
			}()

			// 调用ADB命令清除当前应用
			_, err := adbInstance.ExecuteCommand(deviceId, "shell", "am", "kill-all")
			resultCh <- struct {
				deviceId string
				err      error
			}{deviceId: deviceId, err: err}
		}(deviceId)
	}

	// 收集结果
	for i := 0; i < len(req.DeviceIds); i++ {
		result := <-resultCh
		if result.err != nil {
			results[result.deviceId] = result.err.Error()
		} else {
			results[result.deviceId] = ""
		}
	}

	return results, nil
}
