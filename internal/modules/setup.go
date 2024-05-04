package modules

import (
	"ultraphx-core/internal/modules/alert"
	"ultraphx-core/internal/modules/data"
)

func Setup() {
	data.Setup()
	alert.Setup()
}
