package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Texecom       TexecomConfig       `yaml:"texecom"`
	MQTT          MQTTConfig          `yaml:"mqtt"`
	HomeAssistant HomeAssistantConfig `yaml:"homeassistant"`
	Zones         []ZoneConfig        `yaml:"zones"`
	Areas         []AreaConfig        `yaml:"areas"`
	Log           string              `yaml:"log"`
	Cache         bool                `yaml:"cache"`
}

type TexecomConfig struct {
	Host        string `yaml:"host"`
	UDLPassword string `yaml:"udl_password"`
	Port        int    `yaml:"port"`
}

type MQTTConfig struct {
	ClientID           string `yaml:"client_id"`
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	Keepalive          int    `yaml:"keepalive"`
	Password           string `yaml:"password"`
	QOS                int    `yaml:"qos"`
	Retain             bool   `yaml:"retain"`
	RetainLog          bool   `yaml:"retain_log"`
	Username           string `yaml:"username"`
	CA                 string `yaml:"ca"`
	Cert               string `yaml:"cert"`
	Key                string `yaml:"key"`
	RejectUnauthorized bool   `yaml:"reject_unauthorized"`
	Prefix             string `yaml:"prefix"`
	Clean              bool   `yaml:"clean"`
}

type HomeAssistantConfig struct {
	Discovery bool   `yaml:"discovery"`
	Prefix    string `yaml:"prefix"`
}

type ZoneConfig struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	DeviceClass string `yaml:"device_class"`
}

type AreaConfig struct {
	ID                 string `yaml:"id"`
	Name               string `yaml:"name"`
	Code               string `yaml:"code"`
	CodeArmRequired    bool   `yaml:"code_arm_required"`
	CodeDisarmRequired bool   `yaml:"code_disarm_required"`
	FullArm            string `yaml:"full_arm"`
	PartArm1           string `yaml:"part_arm_1"`
	PartArm2           string `yaml:"part_arm_2"`
	PartArm3           string `yaml:"part_arm_3"`
}

func LoadConfig(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	// Set default values
	if config.MQTT.ClientID == "" {
		config.MQTT.ClientID = "texecom2mqtt"
	}
	if config.MQTT.Host == "" {
		config.MQTT.Host = "localhost"
	}
	if config.MQTT.Port == 0 {
		config.MQTT.Port = 1883
	}
	if config.MQTT.Keepalive == 0 {
		config.MQTT.Keepalive = 60
	}
	if config.MQTT.Prefix == "" {
		config.MQTT.Prefix = "texecom2mqtt"
	}
	if config.HomeAssistant.Prefix == "" {
		config.HomeAssistant.Prefix = "homeassistant"
	}
	if config.Log == "" {
		config.Log = "info"
	}
	if config.Texecom.UDLPassword == "" {
		config.Texecom.UDLPassword = "1234"
	}
	if config.Texecom.Port == 0 {
		config.Texecom.Port = 10001
	}

	return &config, nil
}
