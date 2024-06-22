package router

import (
	"ultraphx-core/internal/api"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine
var apiRouter *gin.RouterGroup
var authRouter *gin.RouterGroup

func init() {
	r = gin.Default()

	// Register middleware
	r.Use(CorsMiddleware)

	apiRouter = r.Group("/api")
	authRouter = apiRouter.Group("/auth")
	authRouter.Use(AuthMiddleware)

	// Register routes
	apiRouter.POST("/plugin/register", api.HandlePluginRegister)
	apiRouter.GET("/plugin/check_active", api.HandlePluginCheckActive)

	authRouter.GET("/client/connected", api.GetConnectedClients)
	authRouter.GET("/client/pending", api.GetPendingClients)
	authRouter.POST("/client/add_active_sensor", api.AddActiveSensor)
	authRouter.POST("/client/remove_client", api.RemoveClient)
	authRouter.POST("/client/set_client_status", api.SetClientStatus)

	authRouter.GET("/system/info", api.GetSystemInfo)
}

func GetApiRouter() *gin.RouterGroup {
	return apiRouter
}

func GetAuthRouter() *gin.RouterGroup {
	return authRouter
}

func Run(addr string) {
	r.Run(addr)
}
