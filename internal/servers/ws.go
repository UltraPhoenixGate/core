package servers

import (
	"net/http"
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
				Type:    "error",
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

func wsHandler(w http.ResponseWriter, r *http.Request, bus *hub.Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade connection to websocket")
		return
	}
	client := hub.NewClient(uuid.New().String(), bus)
	bus.Register(client)
	go readPump(client, conn)
	go writePump(client, conn)
}

func ServeWs(bus *hub.Hub) {
	httpMap := http.NewServeMux()
	httpMap.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, bus) // Pass the hub to the wsHandler function
	})
	logrus.Info("Starting websocket server on :8080")
	http.ListenAndServe(":8080", httpMap)
}
