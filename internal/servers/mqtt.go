package servers

import (
	"fmt"
	"ultraphx-core/internal/hub"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

func MQTTMessageHandler(client mqtt.Client, msg mqtt.Message, h *hub.Hub) {
	message, err := hub.PraseMessageByte(msg.Payload())
	if err != nil {
		logrus.WithError(err).Error("Failed to parse MQTT message")
		return
	}
	h.Broadcast(message)
}

func ConnectMQTTBroker(uri string, h *hub.Hub) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(uri)
	opts.SetClientID("ultraphx-mqtt-client")
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		MQTTMessageHandler(client, msg, h)
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("Failed to connect to MQTT broker")
	}

	return client
}

func SubscribeToTopic(client mqtt.Client, topic string) {
	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal(fmt.Sprintf("Failed to subscribe to topic %s", topic))
	}
}

func ServeMQTT(h *hub.Hub) {
	mqttURI := "tcp://localhost:1883" // replace with your MQTT broker URI
	client := ConnectMQTTBroker(mqttURI, h)
	SubscribeToTopic(client, "#") // subscribe to all topics

	logrus.Info("Connected to MQTT broker and subscribed to topic")
}
