package screenshot

import (
	"context"

	v1 "backend/api/screenshot/v1"
)

type IScreenshotV1 interface {
	Capture(ctx context.Context, req *v1.ScreenshotReq) (res *v1.ScreenshotRes, err error)
}
