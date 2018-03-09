package main

import (
	"github.com/dispatchlabs/disgo/core"
	"github.com/dispatchlabs/disgo_commons/utils"
)

func main() {
	utils.InitMainPackagePath()

	server := core.NewServer()
	server.Go()
}
