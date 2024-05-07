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
	servers.SetupWs(h)
	servers.SetupHttp(h)

	go servers.ServeHTTP(h)
	go servers.ServeMQTT(h)

	// Block forever
	select {}
}

func main() {
	Bootstrap()
}
