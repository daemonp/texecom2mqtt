package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daemonp/texecom2mqtt/internal/panel"
	"github.com/daemonp/texecom2mqtt/internal/types"
)

const cacheFileName = "texecom2mqtt_cache.json"

func SaveCache(p *panel.Panel) error {
	cacheData := types.CacheData{
		Device:     p.GetDevice(),
		Areas:      p.GetAreas(),
		Zones:      p.GetZones(),
		LastUpdate: time.Now(),
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %v", err)
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %v", err)
	}

	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	cacheFilePath := filepath.Join(cacheDir, cacheFileName)
	err = os.WriteFile(cacheFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	return nil
}

func LoadCache() (*types.CacheData, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %v", err)
	}

	cacheFilePath := filepath.Join(cacheDir, cacheFileName)
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Cache file doesn't exist, return nil without error
		}
		return nil, fmt.Errorf("failed to read cache file: %v", err)
	}

	var cacheData types.CacheData
	err = json.Unmarshal(data, &cacheData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %v", err)
	}

	return &cacheData, nil
}

func DeleteCache() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %v", err)
	}

	cacheFilePath := filepath.Join(cacheDir, cacheFileName)
	err = os.Remove(cacheFilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %v", err)
	}

	return nil
}

func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	return filepath.Join(homeDir, ".cache", "texecom2mqtt"), nil
}

