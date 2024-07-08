# texecom2mqtt

texecom2mqtt is a Go application that bridges Texecom alarm systems with MQTT, enabling integration with home automation systems like Home Assistant. It provides real-time updates on the status of your Texecom alarm system and allows you to control it remotely via MQTT messages.

## Features

- Connect to Texecom alarm panels via network connection
- Publish alarm system status to MQTT topics
- Control the alarm system (arm, disarm, reset) via MQTT commands
- Automatic discovery and integration with Home Assistant
- Caching of panel data for faster startup
- Detailed logging for troubleshooting
- Configurable via YAML file

This will create an executable named `texecom2mqtt` in the project directory.

## Configuration

Create a `config.yml` file in the same directory as the executable. Here's an example configuration:

```yaml
texecom:
host: "192.168.1.100"  # IP address of your Texecom panel
udl_password: "1234"   # UDL password for the panel
port: 10001            # Port number (usually 10001)

mqtt:
host: "localhost"      # MQTT broker address
port: 1883             # MQTT broker port
username: ""           # MQTT username (if required)
password: ""           # MQTT password (if required)
client_id: "texecom2mqtt"
prefix: "texecom2mqtt" # MQTT topic prefix
qos: 0                 # MQTT QoS level
retain: true           # Whether to retain MQTT messages
retain_log: false      # Whether to retain log messages

homeassistant:
discovery: true        # Enable Home Assistant MQTT discovery
prefix: "homeassistant"

log: "info"              # Log level (trace, debug, info, warning, error)
cache: true              # Enable caching of panel data

areas:
- id: "A"
 name: "House"
 code: "1234"
 code_arm_required: false
 code_disarm_required: true
 full_arm: "armed_away"
 part_arm_1: "armed_home"

zones:
- id: "1"
 name: "Front Door"
 device_class: "door"
- id: "2"
 name: "Living Room PIR"
 device_class: "motion"
 ```

