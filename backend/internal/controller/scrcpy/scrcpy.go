package scrcpy

import (
	"context"

	v1 "backend/api/scrcpy/v1"
	"backend/internal/service"
)

func (c *ControllerV1) StartStream(ctx context.Context, req *v1.StartStreamReq) (res *v1.StartStreamRes, err error) {
	return service.Scrcpy().StartStream(ctx, req)
}

func (c *ControllerV1) StopStream(ctx context.Context, req *v1.StopStreamReq) (res *v1.StopStreamRes, err error) {
	return service.Scrcpy().StopStream(ctx, req)
}
