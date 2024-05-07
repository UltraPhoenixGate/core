package router

import (
	"github.com/gorilla/mux"
)

var router *mux.Router
var authRouter *mux.Router

func init() {
	r := mux.NewRouter()
	router = r.PathPrefix("/api").Subrouter()
	authRouter = router.PathPrefix("/auth").Subrouter()
	authRouter.Use(AuthMiddleware)
}

func GetRouter() *mux.Router {
	return router
}

func GetAuthRouter() *mux.Router {
	return authRouter
}
