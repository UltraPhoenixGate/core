package servers

import (
	"io"
	"net/http"
	"ultraphx-core/internal/api"
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

func ServeHttp(h *hub.Hub) {
	httpMap := http.NewServeMux()
	httpMap.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		httpBroadcastHandler(w, r, h) // Pass the hub to the httpBroadcastHandler function
	})

	// plugin register
	httpMap.HandleFunc("/plugin/register", api.HandlePluginRegister)
	httpMap.HandleFunc("/plugin/check-active", api.HandlePluginCheckActive)

	logrus.Info("Starting HTTP server on :8081")
	http.ListenAndServe(":8081", httpMap)
}
