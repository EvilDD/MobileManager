package device

import (
	"context"

	v1 "backend/api/device/v1"
)

type IDeviceV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
