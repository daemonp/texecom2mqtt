package panel

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/daemonp/texecom2mqtt/internal/config"
	"github.com/daemonp/texecom2mqtt/internal/log"
	"github.com/daemonp/texecom2mqtt/internal/texecom"
	"github.com/daemonp/texecom2mqtt/internal/types"
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
	p.log.Debug("Attempting connection to %s:%d", p.config.Texecom.Host, p.config.Texecom.Port)
	err := p.texecom.Connect(p.config.Texecom.Host, p.config.Texecom.Port)
	if err != nil {
		p.log.Error("Failed to connect to panel: %v", err)
		return fmt.Errorf("failed to connect to panel: %v", err)
	}
	p.log.Info("Connected to panel")
	return nil
}

func (p *Panel) Login() error {
	p.log.Info("Logging in to panel...")
	p.log.Debug("Sending login command with UDL password")
	err := p.texecom.Login(p.config.Texecom.UDLPassword)
	if err != nil {
		p.log.Error("Failed to log in to panel: %v", err)
		return fmt.Errorf("failed to log in to panel: %v", err)
	}
	p.isLoggedIn = true
	p.log.Info("Successfully logged in to panel")
	return nil
}

func (p *Panel) Start() error {
	if !p.isLoggedIn {
		return fmt.Errorf("not logged in to panel")
	}

	p.log.Info("Starting panel operations...")

	p.log.Debug("Loading initial data from panel")
	if err := p.loadInitialData(); err != nil {
		p.log.Error("Failed to load initial data: %v", err)
		return fmt.Errorf("failed to load initial data: %v", err)
	}

	p.log.Debug("Starting event listener")
	go p.listenForEvents()

	p.log.Debug("Starting keepalive routine")
	go p.keepalive()

	p.log.Info("Panel operations started successfully")
	return nil
}

func (p *Panel) loadInitialData() error {
	var err error

	p.log.Debug("Fetching panel identification")
	p.device, err = p.texecom.GetPanelIdentification()
	if err != nil {
		return fmt.Errorf("failed to get panel identification: %v", err)
	}
	p.log.Debug("Panel identification: %+v", p.device)

	p.log.Debug("Fetching areas")
	p.areas, err = p.texecom.GetAllAreas()
	if err != nil {
		return fmt.Errorf("failed to get areas: %v", err)
	}
	p.log.Debug("Fetched %d areas", len(p.areas))

	p.log.Debug("Fetching zones")
	p.zones, err = p.texecom.GetAllZones()
	if err != nil {
		return fmt.Errorf("failed to get zones: %v", err)
	}
	p.log.Debug("Fetched %d zones", len(p.zones))

	for i, area := range p.areas {
		p.areas[i].Name = normalize(area.Name)
	}

	for i, zone := range p.zones {
		p.zones[i].Name = normalize(zone.Name)
	}

	p.log.Debug("Updating zone states")
	if err := p.updateZoneStates(); err != nil {
		return fmt.Errorf("failed to update zone states: %v", err)
	}

	p.log.Debug("Updating area states")
	if err := p.updateAreaStates(); err != nil {
		return fmt.Errorf("failed to update area states: %v", err)
	}

	p.log.Info("Initial data loaded successfully")
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

func (p *Panel) Arm(area int, armType types.ArmType) error {
	return p.texecom.Arm(area, armType)
}

func (p *Panel) Disarm(area int) error {
	return p.texecom.Disarm(area)
}

func (p *Panel) Reset(area int) error {
	return p.texecom.Reset(area)
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
	return &types.CacheData{
		Device:     p.device,
		Areas:      p.areas,
		Zones:      p.zones,
		LastUpdate: time.Now(),
	}
}

func (p *Panel) Disconnect() {
	p.log.Info("Disconnecting from panel...")
	p.texecom.Disconnect()
	p.log.Info("Disconnected from panel")
}

func normalize(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}
