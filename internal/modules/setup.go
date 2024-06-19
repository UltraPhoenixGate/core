package modules

import (
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/modules/alert"
	"ultraphx-core/internal/modules/camera"
	"ultraphx-core/internal/modules/collect"
	"ultraphx-core/internal/modules/data"
)

func Setup(h *hub.Hub) {
	data.Setup()
	alert.Setup()
	camera.Setup()
	collect.Setup(h)
}
