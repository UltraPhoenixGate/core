package servers

import (
	"net/http"
	"net/url"
	"time"
	"ultraphx-core/internal/hub"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func readPump(client *hub.Client, conn *websocket.Conn) {
	defer func() {
		client.Hub.Unregister(client)
		conn.Close()
	}()

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Error("Failed to read message from websocket")
			break
		}
		msg, err := hub.PraseMessageByte(payload)
		if err != nil {
			client.Send(&hub.Message{
				Topic:   "error",
				Payload: "Failed to parse message",
			})
		}
		client.Hub.Broadcast(msg)
	}
}

func writePump(client *hub.Client, conn *websocket.Conn) {
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.SendChan:
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message.ToJson()); err != nil {
				logrus.WithError(err).Error("Failed to write message to websocket")
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request, h *hub.Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade connection to websocket")
		return
	}
	client := hub.NewClient(uuid.New().String(), h)
	h.Register(client)
	go readPump(client, conn)
	go writePump(client, conn)
}

func ServeWs(h *hub.Hub) {
	httpMap := http.NewServeMux()
	httpMap.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, h) // Pass the hub to the wsHandler function
	})
	// 反向ws
	httpMap.HandleFunc("/ws-reverse", func(w http.ResponseWriter, r *http.Request) {
		wsUrl := r.URL.Query().Get("url")
		if wsUrl == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}
		u, err := url.Parse(wsUrl)
		if err != nil {
			http.Error(w, "url is invalid", http.StatusBadRequest)
			return
		}
		ConnectWs(u, h)
	})
	logrus.Info("Starting websocket server on :8080")
	http.ListenAndServe(":8080", httpMap)
}

func ConnectWs(u *url.URL, h *hub.Hub) {
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to websocket")
		return
	}
	client := hub.NewClient(uuid.New().String(), h)
	h.Register(client)
	go readPump(client, conn)
	go writePump(client, conn)
}
