package cmd

import (
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/servers"
)

func Bootstrap() {
	h := hub.NewHub()
	go h.Run()
	go servers.ServeWs(h)

	// Block forever
	select {}
}
