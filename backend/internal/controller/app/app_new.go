package app

import "backend/api/app"

// ControllerV1 实现了 app.IAppV1 接口
type ControllerV1 struct{}

var _ app.IAppV1 = (*ControllerV1)(nil)

// NewV1 创建 ControllerV1 实例
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}
