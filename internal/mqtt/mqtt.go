package mqtt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/daemonp/texecom2mqtt/internal/config"
	"github.com/daemonp/texecom2mqtt/internal/log"
	"github.com/daemonp/texecom2mqtt/internal/panel"
)

const (
	offlinePayload = "offline"
	onlinePayload  = "online"
)

type MQTT struct {
	config  *config.MQTTConfig
	panel   *panel.Panel
	log     *log.Logger
	client  mqtt.Client
	topics  *Topics
	mu      sync.Mutex
}

func NewMQTT(cfg *config.MQTTConfig, p *panel.Panel, logger *log.Logger) *MQTT {
	return &MQTT{
		config: cfg,
		panel:  p,
		log:    logger,
		topics: NewTopics(cfg.Prefix),
	}
}

func (m *MQTT) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", m.config.Host, m.config.Port))
	opts.SetClientID(m.config.ClientID)
	opts.SetUsername(m.config.Username)
	opts.SetPassword(m.config.Password)
	opts.SetCleanSession(m.config.Clean)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(m.onConnect)
	opts.SetConnectionLostHandler(m.onDisconnect)

	opts.SetWill(m.topics.Status(), offlinePayload, byte(m.config.QOS), m.config.Retain)

	m.client = mqtt.NewClient(opts)

	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	m.log.Info("Connected to MQTT broker: %s:%d", m.config.Host, m.config.Port)
	return nil
}

func (m *MQTT) onConnect(client mqtt.Client) {
	m.log.Info("MQTT connection established")
	m.publishOnlineStatus()
	m.subscribeTopics()
	m.publishPanelStatus()
}

func (m *MQTT) onDisconnect(client mqtt.Client, err error) {
	m.log.Error("MQTT connection lost: %v", err)
}

func (m *MQTT) subscribeTopics() {
	topics := []string{
		m.topics.Text(),
		m.topics.DateTime(),
	}

	for _, area := range m.panel.GetAreas() {
		topics = append(topics, m.topics.AreaCommand(area))
	}

	for _, topic := range topics {
		token := m.client.Subscribe(topic, byte(m.config.QOS), m.handleMessage)
		if token.Wait() && token.Error() != nil {
			m.log.Error("Failed to subscribe to topic %s: %v", topic, token.Error())
		} else {
			m.log.Debug("Subscribed to topic: %s", topic)
		}
	}
}

func (m *MQTT) handleMessage(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())

	m.log.Debug("Received message on topic %s: %s", topic, payload)

	switch topic {
	case m.topics.Text():
		m.panel.SetLCDDisplay(payload)
	case m.topics.DateTime():
		t, err := time.Parse(time.RFC3339, payload)
		if err != nil {
			m.log.Error("Invalid datetime format: %s", payload)
			return
		}
		m.panel.SetDateTime(t)
	default:
		for _, area := range m.panel.GetAreas() {
			if topic == m.topics.AreaCommand(area) {
				m.handleAreaCommand(area, payload)
				return
			}
		}
		m.log.Warning("Received message on unknown topic: %s", topic)
	}
}

func (m *MQTT) handleAreaCommand(area panel.Area, command string) {
	switch command {
	case "full_arm":
		m.panel.Arm(area, panel.ArmTypeFull)
	case "part_arm_1":
		m.panel.Arm(area, panel.ArmTypePartArm1)
	case "part_arm_2":
		m.panel.Arm(area, panel.ArmTypePartArm2)
	case "part_arm_3":
		m.panel.Arm(area, panel.ArmTypePartArm3)
	case "disarm":
		m.panel.Disarm(area)
	default:
		m.log.Warning("Unknown area command: %s", command)
	}
}

func (m *MQTT) publishOnlineStatus() {
	m.publish(m.topics.Status(), onlinePayload, true)
}

func (m *MQTT) publishPanelStatus() {
	device := m.panel.GetDevice()
	status := map[string]interface{}{
		"model":            device.Model,
		"serial_number":    device.SerialNumber,
		"firmware_version": device.FirmwareVersion,
	}
	m.publish(m.topics.Config(), status, true)
}

func (m *MQTT) PublishAreaStatus(area panel.Area) {
	status := map[string]interface{}{
		"id":     area.ID,
		"name":   area.Name,
		"number": area.Number,
		"status": area.Status.String(),
	}
	if area.Status == panel.AreaStatePartArmed {
		status["part_arm"] = area.PartArm
	}
	m.publish(m.topics.Area(area), status, true)
}

func (m *MQTT) PublishZoneStatus(zone panel.Zone) {
	status := map[string]interface{}{
		"id":     zone.ID,
		"name":   zone.Name,
		"number": zone.Number,
		"type":   zone.Type.String(),
		"status": zone.Status.String(),
	}
	m.publish(m.topics.Zone(zone), status, true)
}

func (m *MQTT) PublishLogEvent(event panel.LogEvent) {
	m.publish(m.topics.Log(), event, m.config.RetainLog)
}

func (m *MQTT) publish(topic string, message interface{}, retain bool) {
	payload, err := json.Marshal(message)
	if err != nil {
		m.log.Error("Failed to marshal message for topic %s: %v", topic, err)
		return
	}

	token := m.client.Publish(topic, byte(m.config.QOS), retain, payload)
	if token.Wait() && token.Error() != nil {
		m.log.Error("Failed to publish message to topic %s: %v", topic, token.Error())
	} else {
		m.log.Debug("Published message to topic: %s", topic)
	}
}

func (m *MQTT) Close() {
	if m.client != nil && m.client.IsConnected() {
		m.publish(m.topics.Status(), offlinePayload, true)
		m.client.Disconnect(250)
	}
}
