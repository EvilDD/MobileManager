package service

import (
	v1 "backend/api/scrcpy/v1"
	"context"
)

type IScrcpy interface {
	StartStream(ctx context.Context, req *v1.StartStreamReq) (res *v1.StartStreamRes, err error)
	StopStream(ctx context.Context, req *v1.StopStreamReq) (res *v1.StopStreamRes, err error)
}

var localScrcpy IScrcpy

func Scrcpy() IScrcpy {
	if localScrcpy == nil {
		panic("implement not found for interface IScrcpy")
	}
	return localScrcpy
}
