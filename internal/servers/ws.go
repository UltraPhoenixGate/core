package servers

import (
	"net/http"
	"net/url"
	"time"
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/router"

	"github.com/gin-gonic/gin"
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
		client.Broadcast(msg)
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

func wsHandler(c *gin.Context, h *hub.Hub) {
	// client := c.MustGet("client").(models.Client)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade connection to websocket")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	hubClient := hub.NewClient(uuid.New().String(), h)
	h.Register(hubClient)
	logrus.WithField("client_id", hubClient.ID).Info("Client connected")
	go readPump(hubClient, conn)
	go writePump(hubClient, conn)
}

func SetupWs(h *hub.Hub) {
	authRouter := router.GetAuthRouter()
	authRouter.GET("/ws", func(c *gin.Context) {
		wsHandler(c, h) // Pass the hub to the wsHandler function
	})
	// 反向ws
	authRouter.GET("/ws-reverse", func(c *gin.Context) {
		wsUrl := c.Query("url")
		if wsUrl == "" {
			c.String(http.StatusBadRequest, "url is required")
			return
		}
		u, err := url.Parse(wsUrl)
		if err != nil {
			c.String(http.StatusBadRequest, "url is invalid")
			return
		}
		ConnectWs(u, h)
	})
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
