package scrcpy

import (
	v1 "backend/api/scrcpy/v1"
	"context"
)

type IScrcpyV1 interface {
	StartStream(ctx context.Context, req *v1.StartStreamReq) (res *v1.StartStreamRes, err error)
	StopStream(ctx context.Context, req *v1.StopStreamReq) (res *v1.StopStreamRes, err error)
}
