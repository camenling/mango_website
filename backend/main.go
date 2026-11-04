package main

import (
	_ "mangosign/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"mangosign/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
