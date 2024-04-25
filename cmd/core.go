package cmd

import (
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/servers"
)

func Bootstrap() {
	globalHub := hub.NewHub()
	go globalHub.Run()
	go servers.ServeWs(globalHub)

	// Block forever
	select {}
}
