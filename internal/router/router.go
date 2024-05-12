package router

import (
	"net/http"
	"ultraphx-core/internal/api"

	"github.com/gorilla/mux"
)

var apiRouter *mux.Router
var authRouter *mux.Router

func init() {
	r := mux.NewRouter()
	apiRouter = r.PathPrefix("/api").Subrouter()
	authRouter = apiRouter.PathPrefix("/auth").Subrouter()
	authRouter.Use(AuthMiddleware)

	// Register routes
	apiRouter.HandleFunc("/plugin/register", api.HandlePluginRegister).Methods(http.MethodPost)
	apiRouter.HandleFunc("/plugin/check_active", api.HandlePluginCheckActive).Methods(http.MethodGet)
}

func GetRouter() *mux.Router {
	return apiRouter
}

func GetAuthRouter() *mux.Router {
	return authRouter
}
