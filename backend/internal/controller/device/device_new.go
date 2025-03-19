package device

import (
	"backend/api/device"
)

type ControllerV1 struct{}

func NewV1() device.IDeviceV1 {
	return &ControllerV1{}
}
