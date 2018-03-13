package main

import (
	"github.com/dispatchlabs/disgo/core"
	"github.com/dispatchlabs/commons/utils"
)

func main() {
	utils.InitMainPackagePath()

	server := core.NewServer()
	server.Go()
}
