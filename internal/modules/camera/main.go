package camera

import (
	"ultraphx-core/internal/models"
	"ultraphx-core/internal/router"
)

func Setup() {
	models.AutoMigrate(&Camera{})

	authRouter := router.GetAuthRouter()
	authRouter.POST("/camera", AddCamera)
	authRouter.GET("/cameras", GetCameras)
	authRouter.DELETE("/camera", DeleteCamera)
	authRouter.PUT("/camera", UpdateCamera)

	authRouter.GET("/camera/capture", GetCurrentFrame)
	authRouter.GET("/camera/stream", OpenStream)
	authRouter.GET("/camera/onvif/scan", ScanOnvifDevices)
}
