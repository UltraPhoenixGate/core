package cmd

import (
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/servers"
)

func Bootstrap() {
	h := hub.NewHub()
	go h.Run()
	go servers.ServeWs(h)
	go servers.ServeHttp(h)
	go servers.ServeMQTT(h)

	// Block forever
	select {}
}
