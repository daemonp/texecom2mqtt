package homeassistant

import (
	"strings"

	"github.com/daemonp/texecom2mqtt/internal/types"
)

func getDeviceClass(zone types.Zone) string {
	// Check if there's a custom device class set in the config
	if zone.HomeAssistant != nil && zone.HomeAssistant.DeviceClass != "" {
		return zone.HomeAssistant.DeviceClass
	}

	// Try to guess the device class based on the zone name
	name := strings.ToLower(zone.Name)
	if strings.Contains(name, "pir") {
		return "motion"
	}
	if strings.Contains(name, "door") {
		return "door"
	}
	if strings.Contains(name, "window") {
		return "window"
	}
	if strings.Contains(name, "smoke") || strings.Contains(name, "fire") {
		return "smoke"
	}
	if strings.Contains(name, "gas") {
		return "gas"
	}
	if strings.Contains(name, "water") {
		return "moisture"
	}

	// Default to motion if we can't determine a more specific type
	return "motion"
}
