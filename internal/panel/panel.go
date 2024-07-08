package panel

import (
	"fmt"
	"sync"
	"time"

	"github.com/daemonp/texecom2mqtt/internal/config"
	"github.com/daemonp/texecom2mqtt/internal/log"
	"github.com/daemonp/texecom2mqtt/internal/texecom"
	"github.com/daemonp/texecom2mqtt/internal/types"
	"github.com/daemonp/texecom2mqtt/internal/util"
)

type Panel struct {
	config     *config.Config
	log        *log.Logger
	texecom    *texecom.Texecom
	areas      []types.Area
	zones      []types.Zone
	device     types.Device
	mu         sync.Mutex
	isLoggedIn bool
}

func NewPanel(cfg *config.Config, logger *log.Logger) *Panel {
	return &Panel{
		config:  cfg,
		log:     logger,
		texecom: texecom.NewTexecom(logger),
	}
}

func (p *Panel) Connect() error {
	p.log.Info("Connecting to panel...")
	err := p.texecom.Connect(p.config.Texecom.Host, p.config.Texecom.Port)
	if err != nil {
		return fmt.Errorf("failed to connect to panel: %v", err)
	}
	p.log.Info("Connected to panel")
	return nil
}

func (p *Panel) Login() error {
	p.log.Info("Logging in to panel...")
	err := p.texecom.Login(p.config.Texecom.UDLPassword)
	if err != nil {
		return fmt.Errorf("failed to log in to panel: %v", err)
	}
	p.isLoggedIn = true
	p.log.Info("Logged in to panel")
	return nil
}

func (p *Panel) Start() error {
	if !p.isLoggedIn {
		return fmt.Errorf("not logged in to panel")
	}

	p.log.Info("Starting panel operations...")

	if err := p.loadInitialData(); err != nil {
		return fmt.Errorf("failed to load initial data: %v", err)
	}

	// Start event listening
	go p.listenForEvents()

	// Start keepalive
	go p.keepalive()

	return nil
}

func (p *Panel) loadInitialData() error {
	var err error
	p.device, err = p.texecom.GetPanelIdentification()
	if err != nil {
		return fmt.Errorf("failed to get panel identification: %v", err)
	}

	p.areas, err = p.texecom.GetAllAreas()
	if err != nil {
		return fmt.Errorf("failed to get areas: %v", err)
	}

	p.zones, err = p.texecom.GetAllZones()
	if err != nil {
		return fmt.Errorf("failed to get zones: %v", err)
	}

	for i, area := range p.areas {
		p.areas[i].Name = util.Normalize(area.Name)
	}

	for i, zone := range p.zones {
		p.zones[i].Name = util.Normalize(zone.Name)
	}

	if err := p.updateZoneStates(); err != nil {
		return fmt.Errorf("failed to update zone states: %v", err)
	}

	if err := p.updateAreaStates(); err != nil {
		return fmt.Errorf("failed to update area states: %v", err)
	}

	p.log.Info("Initial data loaded")
	return nil
}

func (p *Panel) listenForEvents() {
	for event := range p.texecom.Events() {
		p.handleEvent(event)
	}
}

func (p *Panel) handleEvent(event interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch e := event.(type) {
	case types.ZoneEvent:
		p.handleZoneEvent(e)
	case types.AreaEvent:
		p.handleAreaEvent(e)
	case types.LogEvent:
		p.handleLogEvent(e)
	}
}

func (p *Panel) handleZoneEvent(event types.ZoneEvent) {
	for i, zone := range p.zones {
		if zone.Number == event.ZoneNumber {
			p.zones[i].Status = event.ZoneState
			p.log.Info("Zone %s (%d) status changed to %s", zone.Name, zone.Number, event.ZoneState)
			break
		}
	}
}

func (p *Panel) handleAreaEvent(event types.AreaEvent) {
	for i, area := range p.areas {
		if area.Number == event.AreaNumber {
			p.areas[i].Status = event.AreaState
			if event.AreaState == types.AreaStatePartArmed {
				p.areas[i].PartArm = event.PartArm
			}
			p.log.Info("Area %s (%d) status changed to %s", area.Name, area.Number, event.AreaState)
			break
		}
	}
}

func (p *Panel) handleLogEvent(event types.LogEvent) {
	p.log.Panel("Log event: %s", event.Description)
}

func (p *Panel) keepalive() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		if err := p.texecom.UpdateSystemPower(); err != nil {
			p.log.Error("Failed to update system power: %v", err)
		}
	}
}

func (p *Panel) updateZoneStates() error {
	states, err := p.texecom.GetZoneStates()
	if err != nil {
		return err
	}

	for i, state := range states {
		if i < len(p.zones) {
			p.zones[i].Status = state
		}
	}

	return nil
}

func (p *Panel) updateAreaStates() error {
	states, err := p.texecom.GetAreaStates()
	if err != nil {
		return err
	}

	for i, state := range states {
		if i < len(p.areas) {
			p.areas[i].Status = state.Status
			p.areas[i].PartArm = state.PartArm
		}
	}

	return nil
}

func (p *Panel) Arm(area types.Area, armType types.ArmType) error {
	return p.texecom.Arm(area.Number, armType)
}

func (p *Panel) Disarm(area types.Area) error {
	return p.texecom.Disarm(area.Number)
}

func (p *Panel) Reset(area types.Area) error {
	return p.texecom.Reset(area.Number)
}

func (p *Panel) SetDateTime(t time.Time) error {
	return p.texecom.SetDateTime(t)
}

func (p *Panel) SetLCDDisplay(text string) error {
	return p.texecom.SetLCDDisplay(text)
}

func (p *Panel) GetAreas() []types.Area {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.areas
}

func (p *Panel) GetZones() []types.Zone {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.zones
}

func (p *Panel) GetDevice() types.Device {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.device
}

func (p *Panel) SetCachedData(data *types.CacheData) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.device = data.Device
	p.areas = data.Areas
	p.zones = data.Zones
}

func (p *Panel) GetCacheableData() *types.CacheData {
	p.mu.Lock()
	defer p.mu.Unlock()
	return &types.CacheData{
		Device: p.device,
		Areas:  p.areas,
		Zones:  p.zones,
	}
}

func (p *Panel) Disconnect() {
	p.log.Info("Disconnecting from panel...")
	p.texecom.Disconnect()
	p.log.Info("Disconnected from panel")
}

