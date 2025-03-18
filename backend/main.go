package main

import (
	_ "backend/internal/boot"
	_ "backend/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"backend/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.New())
}
