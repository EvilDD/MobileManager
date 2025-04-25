package boot

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

// 初始化应用表
func initAppTable(ctx context.Context) error {
	_, err := g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "app" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"name" VARCHAR(255) NOT NULL,
			"package_name" VARCHAR(255) NOT NULL,
			"version" VARCHAR(50) NOT NULL,
			"size" INTEGER NOT NULL,
			"app_type" VARCHAR(50) NOT NULL,
			"apk_path" VARCHAR(255) NOT NULL,
			"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE("package_name", "version")
		)
	`)
	return err
}

// 初始化文件表
func initFileTable(ctx context.Context) error {
	_, err := g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "file" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"name" VARCHAR(255) NOT NULL,
			"original_name" VARCHAR(255) NOT NULL,
			"file_type" VARCHAR(50) NOT NULL,
			"file_size" INTEGER NOT NULL,
			"file_path" VARCHAR(255) NOT NULL,
			"mime_type" VARCHAR(100),
			"md5" VARCHAR(32),
			"status" INTEGER DEFAULT 1,
			"created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// 初始化数据表
func initTables(ctx context.Context) error {
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
			"group_id" INTEGER DEFAULT 0,
			"created_at" DATETIME,
			"updated_at" DATETIME,
			FOREIGN KEY ("group_id") REFERENCES "group" ("id")
		)
	`)
	if err != nil {
		return err
	}

	// 创建应用表
	if err := initAppTable(ctx); err != nil {
		return err
	}

	// 创建文件表
	if err := initFileTable(ctx); err != nil {
		return err
	}

	return nil
}

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
