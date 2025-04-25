package file

import "backend/api/file"

// ControllerV1 实现了 file.IFileV1 接口
type ControllerV1 struct{}

var _ file.IFileV1 = (*ControllerV1)(nil)

// NewV1 创建 ControllerV1 实例
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}
