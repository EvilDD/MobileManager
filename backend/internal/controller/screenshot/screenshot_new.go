package screenshot

import (
	"backend/api/screenshot"
)

type ControllerV1 struct{}

func NewV1() screenshot.IScreenshotV1 {
	return &ControllerV1{}
}
