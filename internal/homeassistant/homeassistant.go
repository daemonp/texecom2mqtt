package homeassistant

import (
	"encoding/json"
	"fmt"

	"github.com/daemonp/texecom2mqtt/internal/config"
	"github.com/daemonp/texecom2mqtt/internal/log"
	"github.com/daemonp/texecom2mqtt/internal/mqtt"
	"github.com/daemonp/texecom2mqtt/internal/panel"
	"github.com/daemonp/texecom2mqtt/internal/types"
	"github.com/daemonp/texecom2mqtt/internal/util"
)

type HomeAssistant struct {
	config *config.HomeAssistantConfig
	mqtt   mqtt.MQTTClient
	panel  *panel.Panel
	log    *log.Logger
}

type MQTTClient interface {
	GetPrefix() string
	Topics() *mqtt.Topics
	Publish(topic string, payload interface{}, retain bool)
}

func New(cfg *config.HomeAssistantConfig, mqttClient mqtt.MQTTClient, p *panel.Panel, logger *log.Logger) *HomeAssistant {
	return &HomeAssistant{
		config: cfg,
		mqtt:   mqttClient,
		panel:  p,
		log:    logger,
	}
}

func (ha *HomeAssistant) Start() {
	ha.log.Info("Starting Home Assistant integration")
	ha.publishDiscoveryConfig()
}

func (ha *HomeAssistant) publishDiscoveryConfig() {
	ha.publishPanelConfig()

	for _, area := range ha.panel.GetAreas() {
		ha.publishAreaConfig(area)
	}

	for _, zone := range ha.panel.GetZones() {
		ha.publishZoneConfig(zone)
	}
}

func (ha *HomeAssistant) publishPanelConfig() {
	device := ha.panel.GetDevice()
	config := map[string]interface{}{
		"name":         fmt.Sprintf("Texecom %s", device.Model),
		"identifiers":  []string{device.SerialNumber},
		"manufacturer": "Texecom",
		"model":        device.Model,
		"sw_version":   device.FirmwareVersion,
	}

	ha.publishConfig("binary_sensor", "panel", "connectivity", config)
}

func (ha *HomeAssistant) publishAreaConfig(area types.Area) {
	config := map[string]interface{}{
		"name":             area.Name,
		"unique_id":        fmt.Sprintf("%s_area_%s", ha.mqtt.GetPrefix(), util.Slugify(area.Name)),
		"state_topic":      ha.mqtt.Topics().Area(area),
		"command_topic":    ha.mqtt.Topics().AreaCommand(area),
		"payload_disarm":   "disarm",
		"payload_arm_home": "part_arm_1",
		"payload_arm_away": "full_arm",
		"device_class":     "alarm_control_panel",
		"value_template":   "{{ value_json.status }}",
	}

	ha.publishConfig("alarm_control_panel", area.ID, "", config)
}

func (ha *HomeAssistant) publishZoneConfig(zone types.Zone) {
	config := map[string]interface{}{
		"name":           zone.Name,
		"unique_id":      fmt.Sprintf("%s_zone_%s", ha.mqtt.GetPrefix(), util.Slugify(zone.Name)),
		"state_topic":    ha.mqtt.Topics().Zone(zone),
		"device_class":   getDeviceClass(zone),
		"value_template": "{{ value_json.status }}",
		"payload_on":     "Active",
		"payload_off":    "Secure",
	}

	ha.publishConfig("binary_sensor", zone.ID, "", config)
}

func (ha *HomeAssistant) publishConfig(component, objectId, deviceClass string, config map[string]interface{}) {
	topic := fmt.Sprintf("%s/%s/%s/%s/config", ha.config.Prefix, component, ha.mqtt.GetPrefix(), objectId)

	if deviceClass != "" {
		config["device_class"] = deviceClass
	}

	payload, err := json.Marshal(config)
	if err != nil {
		ha.log.Error("Failed to marshal Home Assistant config: %v", err)
		return
	}

	ha.mqtt.Publish(topic, string(payload), true)
}
