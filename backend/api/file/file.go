package file

import (
	"context"

	v1 "backend/api/file/v1"
)

// IFileV1 文件管理接口
type IFileV1 interface {
	Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error)
	BatchPushByDevices(ctx context.Context, req *v1.BatchPushByDevicesReq) (res *v1.BatchPushByDevicesRes, err error)
	BatchTaskStatus(ctx context.Context, req *v1.BatchTaskStatusReq) (res *v1.BatchTaskStatusRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
}
