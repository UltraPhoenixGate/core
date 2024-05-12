package servers

import (
	"io"
	"net/http"
	"ultraphx-core/internal/api/router"
	"ultraphx-core/internal/config"
	"ultraphx-core/internal/hub"

	"github.com/sirupsen/logrus"
)

func httpBroadcastHandler(w http.ResponseWriter, r *http.Request, h *hub.Hub) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	msg, err := hub.PraseMessageByte(body)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse message")
		http.Error(w, "Failed to parse message", http.StatusBadRequest)
		return
	}

	h.Broadcast(msg)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func SetupHttp(h *hub.Hub) {
	authRouter := router.GetAuthRouter()
	authRouter.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		httpBroadcastHandler(w, r, h) // Pass the hub to the httpBroadcastHandler function
	})
}

func ServeHTTP(h *hub.Hub) {
	httpCfg := config.GetServerConfig()
	logrus.Info("Starting HTTP server on :" + httpCfg.HttpPort)
	http.ListenAndServe(":"+httpCfg.HttpPort, router.GetRouter())
}
