package texecom

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/daemonp/texecom2mqtt/internal/log"
	"github.com/daemonp/texecom2mqtt/internal/types"
)

type Texecom struct {
	log            *log.Logger
	conn           net.Conn
	device         types.Device
	areas          []types.Area
	zones          []types.Zone
	isLoggedIn     bool
	mu             sync.Mutex
	sequence       uint8
	eventChan      chan interface{}
	isConnected    bool
	disconnectChan chan struct{}
}

func NewTexecom(logger *log.Logger) *Texecom {
	return &Texecom{
		log:            logger,
		eventChan:      make(chan interface{}, 100),
		disconnectChan: make(chan struct{}),
	}
}

func (t *Texecom) Connect(host string, port int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isConnected {
		return fmt.Errorf("already connected")
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	t.conn = conn
	t.isConnected = true
	go t.readLoop()

	return nil
}

func (t *Texecom) Disconnect() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return
	}

	close(t.disconnectChan)
	t.conn.Close()
	t.isConnected = false
	close(t.eventChan)
}

func (t *Texecom) Login(password string) error {
	cmd := []byte{0x01} // Login command
	cmd = append(cmd, []byte(password)...)
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		return fmt.Errorf("login failed: invalid response")
	}

	t.isLoggedIn = true
	return nil
}

func (t *Texecom) GetPanelIdentification() (types.Device, error) {
	cmd := []byte{0x16} // Get Panel Identification command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return types.Device{}, fmt.Errorf("failed to get panel identification: %v", err)
	}

	// Parse the response to extract device information
	device := types.Device{
		Model:           string(resp[:20]),
		SerialNumber:    string(resp[20:40]),
		FirmwareVersion: string(resp[40:60]),
		Zones:           int(binary.LittleEndian.Uint16(resp[60:62])),
	}

	t.device = device
	return device, nil
}

func (t *Texecom) GetAllAreas() ([]types.Area, error) {
	cmd := []byte{0x22} // Get Area Text command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get areas: %v", err)
	}

	var areas []types.Area
	for i := 0; i < len(resp); i += 16 {
		areaNumber := i/16 + 1
		areaName := string(resp[i : i+16])
		areas = append(areas, types.Area{
			Number: areaNumber,
			Name:   areaName,
			ID:     fmt.Sprintf("A%d", areaNumber),
		})
	}

	t.areas = areas
	return areas, nil
}

func (t *Texecom) GetAllZones() ([]types.Zone, error) {
	cmd := []byte{0x03} // Get Zone Details command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %v", err)
	}

	var zones []types.Zone
	for i := 0; i < len(resp); i += 32 {
		zoneNumber := i/32 + 1
		zoneName := string(resp[i : i+16])
		zoneType := types.ZoneType(resp[i+16])
		zones = append(zones, types.Zone{
			Number: zoneNumber,
			Name:   zoneName,
			Type:   zoneType,
			ID:     fmt.Sprintf("Z%d", zoneNumber),
		})
	}

	t.zones = zones
	return zones, nil
}

func (t *Texecom) GetZoneStates() ([]types.ZoneState, error) {
	cmd := []byte{0x02} // Get Zone State command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone states: %v", err)
	}

	var states []types.ZoneState
	for _, b := range resp {
		states = append(states, types.ZoneState(b&0x03))
	}

	return states, nil
}

func (t *Texecom) GetAreaStates() ([]types.AreaStatus, error) {
	cmd := []byte{0x0B} // Get Area Flags command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get area states: %v", err)
	}

	var states []types.AreaStatus
	for i := 0; i < len(resp); i += 8 {
		flags := binary.LittleEndian.Uint64(resp[i : i+8])
		status := types.AreaStatus{
			Status:  t.parseAreaState(flags),
			PartArm: t.parsePartArm(flags),
		}
		states = append(states, status)
	}

	return states, nil
}

func (t *Texecom) Arm(areaNumber int, armType types.ArmType) error {
	cmd := []byte{0x06, byte(areaNumber), byte(armType)} // Arm Area command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to arm area: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		return fmt.Errorf("failed to arm area: invalid response")
	}

	return nil
}

func (t *Texecom) Disarm(areaNumber int) error {
	cmd := []byte{0x08, byte(areaNumber)} // Disarm Area command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to disarm area: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		return fmt.Errorf("failed to disarm area: invalid response")
	}

	return nil
}

func (t *Texecom) Reset(areaNumber int) error {
	cmd := []byte{0x09, byte(areaNumber)} // Reset Area command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to reset area: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		return fmt.Errorf("failed to reset area: invalid response")
	}

	return nil
}

func (t *Texecom) SetDateTime(datetime time.Time) error {
	cmd := []byte{0x18} // Set Date/Time command
	cmd = append(cmd, byte(datetime.Day()), byte(datetime.Month()), byte(datetime.Year()%100),
		byte(datetime.Hour()), byte(datetime.Minute()), byte(datetime.Second()))
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to set date/time: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		return fmt.Errorf("failed to set date/time: invalid response")
	}

	return nil
}

func (t *Texecom) SetLCDDisplay(text string) error {
	if len(text) > 32 {
		text = text[:32]
	}
	cmd := []byte{0x0E} // Set LCD Display command
	cmd = append(cmd, []byte(text)...)
	cmd = append(cmd, make([]byte, 32-len(text))...) // Pad with spaces
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to set LCD display: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		return fmt.Errorf("failed to set LCD display: invalid response")
	}

	return nil
}

func (t *Texecom) UpdateSystemPower() error {
	cmd := []byte{0x19} // Get System Power command
	_, err := t.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to update system power: %v", err)
	}

	return nil
}

func (t *Texecom) Events() <-chan interface{} {
	return t.eventChan
}

func (t *Texecom) sendCommand(cmd []byte) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return nil, fmt.Errorf("not connected")
	}

	t.sequence++
	header := []byte{'t', 'C', byte(len(cmd) + 6), t.sequence}
	packet := append(header, cmd...)
	crc := t.calculateCRC(packet)
	packet = append(packet, crc)

	_, err := t.conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	resp := make([]byte, 1024)
	n, err := t.conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return resp[:n], nil
}

func (t *Texecom) readLoop() {
	buffer := make([]byte, 1024)
	for {
		select {
		case <-t.disconnectChan:
			return
		default:
			n, err := t.conn.Read(buffer)
			if err != nil {
				t.log.Error("Read error: %v", err)
				t.Disconnect()
				return
			}

			t.processMessage(buffer[:n])
		}
	}
}

func (t *Texecom) processMessage(msg []byte) {
	if len(msg) < 5 {
		return
	}

	switch msg[1] {
	case 'M': // Event message
		event := t.parseEvent(msg[4:])
		t.eventChan <- event
	case 'R': // Response message
		// Handle responses if needed
	}
}

func (t *Texecom) parseEvent(data []byte) interface{} {
	if len(data) < 2 {
		return nil
	}

	eventType := data[0]
	switch eventType {
	case 1: // Zone event
		return t.parseZoneEvent(data[1:])
	case 2: // Area event
		return t.parseAreaEvent(data[1:])
	case 5: // Log event
		return t.parseLogEvent(data[1:])
	default:
		return nil
	}
}

func (t *Texecom) parseZoneEvent(data []byte) types.ZoneEvent {
	return types.ZoneEvent{
		ZoneNumber: int(binary.LittleEndian.Uint16(data[:2])),
		ZoneState:  types.ZoneState(data[2] & 0x03),
	}
}

func (t *Texecom) parseAreaEvent(data []byte) types.AreaEvent {
	return types.AreaEvent{
		AreaNumber: int(data[0]),
		AreaState:  types.AreaState(data[1]),
	}
}

func (t *Texecom) parseLogEvent(data []byte) types.LogEvent {
	return types.LogEvent{
		Type:        types.LogEventType(data[0]),
		GroupType:   types.LogEventGroupType(data[1]),
		Parameter:   binary.LittleEndian.Uint16(data[2:4]),
		Areas:       binary.LittleEndian.Uint16(data[4:6]),
		Time:        t.parseTimestamp(data[6:10]),
		Description: t.getLogEventDescription(types.LogEventType(data[0])),
	}
}

func (t *Texecom) parseTimestamp(data []byte) time.Time {
	timestamp := binary.LittleEndian.Uint32(data)
	seconds := timestamp & 63
	minutes := (timestamp >> 6) & 63
	hours := (timestamp >> 12) & 31
	day := (timestamp >> 17) & 31
	month := (timestamp >> 22) & 15
	year := 2000 + ((timestamp >> 26) & 63)

	return time.Date(int(year), time.Month(month), int(day), int(hours), int(minutes), int(seconds), 0, time.UTC)
}

func (t *Texecom) parseAreaState(flags uint64) types.AreaState {
	if flags&(1<<0) != 0 {
		return types.AreaStateInAlarm
	}
	if flags&(1<<21) != 0 || flags&(1<<22) != 0 || flags&(1<<23) != 0 {
		return types.AreaStateArmed
	}
	return types.AreaStateDisarmed
}

func (t *Texecom) parsePartArm(flags uint64) int {
	if flags&(1<<50) != 0 {
		return 1
	}
	if flags&(1<<51) != 0 {
		return 2
	}
	if flags&(1<<52) != 0 {
		return 3
	}
	return 0
}

func (t *Texecom) calculateCRC(data []byte) byte {
	crc := byte(0)
	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x85
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func (t *Texecom) getLogEventDescription(eventType types.LogEventType) string {
	switch eventType {
	case types.LogEventTypeEntryExit1:
		return "Entry/Exit 1"
	case types.LogEventTypeEntryExit2:
		return "Entry/Exit 2"
	case types.LogEventTypeGuard:
		return "Guard"
	case types.LogEventTypeGuardAccess:
		return "Guard Access"
	case types.LogEventTwentyFourHourAudible:
		return "24hr Audible"
	case types.LogEventTwentyFourHourSilent:
		return "24hr Silent"
	case types.LogEventPAAudible:
		return "PA Audible"
	case types.LogEventPASilent:
		return "PA Silent"
	case types.LogEventFire:
		return "Fire Alarm"
	case types.LogEventMedical:
		return "Medical"
	case types.LogEventTwentyFourHourGas:
		return "24Hr Gas Alarm"
	case types.LogEventAuxiliary:
		return "Auxiliary Alarm"
	case types.LogEventTamper:
		return "24hr Tamper Alarm"
	case types.LogEventExitTerminator:
		return "Exit Terminator"
	case types.LogEventMomentKey:
		return "Keyswitch - Momentary"
	case types.LogEventLatchKey:
		return "Keyswitch - Latching"
	case types.LogEventSecurity:
		return "Security Key"
	case types.LogEventOmitKey:
		return "Omit Key"
	case types.LogEventCustom:
		return "Custom Alarm"
	default:
		return fmt.Sprintf("Unknown Log Event Type: %d", eventType)
	}
}
