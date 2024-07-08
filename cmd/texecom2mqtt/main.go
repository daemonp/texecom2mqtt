package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/daemonp/texecom2mqtt/internal/config"
    "github.com/daemonp/texecom2mqtt/internal/homeassistant"
    "github.com/daemonp/texecom2mqtt/internal/log"
    "github.com/daemonp/texecom2mqtt/internal/mqtt"
    "github.com/daemonp/texecom2mqtt/internal/panel"
    "github.com/daemonp/texecom2mqtt/internal/cache"
)

func main() {
    configFile := flag.String("config", "config.yml", "Path to configuration file")
    flag.Parse()

    // Load configuration
    cfg, err := config.LoadConfig(*configFile)
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Create logger
    logger := log.NewLogger(cfg.Log)

    // Create panel
    p := panel.NewPanel(cfg, logger)

    // Create MQTT client
    mqttClient := mqtt.NewMQTT(&cfg.MQTT, p, logger)

    // Setup graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Connect to panel
    if err := p.Connect(); err != nil {
        logger.Error("Failed to connect to panel: %v", err)
        os.Exit(1)
    }

    // Login to panel
    if err := p.Login(); err != nil {
        logger.Error("Failed to login to panel: %v", err)
        p.Disconnect()
        os.Exit(1)
    }

    // Load cache if enabled
    if cfg.Cache {
        cacheData, err := cache.LoadCache()
        if err != nil {
            logger.Warning("Failed to load cache: %v", err)
        } else if cacheData != nil {
            // Use cached data to initialize panel
            p.SetCachedData(cacheData)
            logger.Info("Loaded data from cache")
        }
    }

    // Start panel operations
    if err := p.Start(); err != nil {
        logger.Error("Failed to start panel operations: %v", err)
        mqttClient.Close()
        p.Disconnect()
        os.Exit(1)
    }

    // Save cache if enabled
    if cfg.Cache {
        cacheData := p.GetCacheableData()
        if err := cache.SaveCache(cacheData); err != nil {
            logger.Warning("Failed to save cache: %v", err)
        } else {
            logger.Info("Saved data to cache")
        }
    }

    // Connect to MQTT broker
    if err := mqttClient.Connect(); err != nil {
        logger.Error("Failed to connect to MQTT broker: %v", err)
        p.Disconnect()
        os.Exit(1)
    }

    // Initialize and start Home Assistant integration if enabled
    if cfg.HomeAssistant.Discovery {
        ha := homeassistant.New(&cfg.HomeAssistant, mqttClient, p, logger)
        ha.Start()
    }

    // Wait for termination signal
    <-sigChan

    // Graceful shutdown
    logger.Info("Shutting down...")
    mqttClient.Close()
    p.Disconnect()

    // Delete cache if enabled
    if cfg.Cache {
        if err := cache.DeleteCache(); err != nil {
            logger.Warning("Failed to delete cache: %v", err)
        } else {
            logger.Info("Deleted cache")
        }
    }
}
