package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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

func Start() {
	// Start the API server
	logrus.Info("Starting API server on :8080")
	http.ListenAndServe(":8080", router)
}
