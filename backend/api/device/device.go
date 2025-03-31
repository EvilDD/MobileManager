package device

import (
	"context"

	v1 "backend/api/device/v1"
)

type IDeviceV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	// 批量回到主菜单
	BatchGoHome(ctx context.Context, req *v1.BatchGoHomeReq) (res *v1.BatchGoHomeRes, err error)
	// 批量清除当前应用
	BatchKillApps(ctx context.Context, req *v1.BatchKillAppsReq) (res *v1.BatchKillAppsRes, err error)
}
