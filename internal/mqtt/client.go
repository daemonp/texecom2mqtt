package mqtt

type MQTTClient interface {
	GetPrefix() string
	Topics() *Topics
	Publish(topic string, payload interface{}, retain bool)
}
