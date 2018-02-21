package main

import (
	"github.com/dispatchlabs/disgo/core"
)

func main() {
	server := core.NewServer()
	server.Go()
}
