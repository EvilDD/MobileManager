package model

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// App 应用模型
type App struct {
	Id          int64       `json:"id"`          // 应用ID
	Name        string      `json:"name"`        // 应用名称
	PackageName string      `json:"packageName"` // 应用包名
	Version     string      `json:"version"`     // 版本号
	Size        int64       `json:"size"`        // 应用大小(KB)
	AppType     string      `json:"appType"`     // 应用类型
	ApkPath     string      `json:"apkPath"`     // APK文件路径
	CreatedAt   *gtime.Time `json:"createdAt"`   // 创建时间
	UpdatedAt   *gtime.Time `json:"updatedAt"`   // 更新时间
}
