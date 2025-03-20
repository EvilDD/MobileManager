package scrcpy

// ControllerV1 串流控制器
type ControllerV1 struct{}

// NewV1 创建V1版本控制器
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}
