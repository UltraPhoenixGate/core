package main

import (
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/modules"
	"ultraphx-core/internal/servers"
)

func Bootstrap() {
	h := hub.NewHub()
	go h.Run()

	// Start all modules
	modules.Setup()

	// Start all servers
	go servers.ServeWs(h)
	go servers.ServeHttp(h)
	go servers.ServeMQTT(h)

	// Block forever
	select {}
}

func main() {
	Bootstrap()
}
