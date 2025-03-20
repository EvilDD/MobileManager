package screenshot

import (
	"context"

	v1 "backend/api/screenshot/v1"
	"backend/internal/service"
)

// Capture 批量设备截图
func (c *ControllerV1) Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error) {
	return service.Screenshot().Capture(ctx, req)
}
