package boot

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

// 初始化函数
func init() {
	ctx := gctx.New()
	// 确保数据目录存在
	if !gfile.Exists("./data") {
		err := gfile.Mkdir("./data")
		if err != nil {
			g.Log().Fatal(ctx, "Failed to create data directory:", err)
		}
	}

	// 初始化数据表
	err := initTables(ctx)
	if err != nil {
		g.Log().Fatal(ctx, "Failed to initialize database tables:", err)
	}
}

// 初始化数据表
func initTables(ctx gctx.Ctx) error {
	// 创建分组表
	_, err := g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "group" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"name" VARCHAR(100) NOT NULL,
			"description" TEXT,
			"created_at" DATETIME,
			"updated_at" DATETIME
		)
	`)
	if err != nil {
		return err
	}

	// 创建设备表
	_, err = g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "device" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"name" VARCHAR(100) NOT NULL,
			"device_id" VARCHAR(100) NOT NULL UNIQUE,
			"status" VARCHAR(20) DEFAULT 'offline',
			"group_id" INTEGER,
			"created_at" DATETIME,
			"updated_at" DATETIME,
			FOREIGN KEY ("group_id") REFERENCES "group" ("id")
		)
	`)
	return err
}
