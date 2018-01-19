package main

import (
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/dispatchlabs/disgo/server"
)

func main() {

	// Setup log.
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Start server.
	server := server.NewServer()
	server.Start()
}