package mqtt

import (
	"fmt"
	"strings"
)

func ParseURL(urlStr string) (string, int) {
	urlStr = strings.TrimPrefix(urlStr, "mqtt://")
	parts := strings.Split(urlStr, ":")
	if len(parts) == 1 {
		return parts[0], 1883 // Default MQTT port
	}
	port := 1883
	fmt.Sscanf(parts[1], "%d", &port)
	return parts[0], port
}
