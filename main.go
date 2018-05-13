package main

import (
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/disgo/core"
)

func main() {
	utils.InitMainPackagePath()

	server := core.NewServer()
	server.Go()
}
