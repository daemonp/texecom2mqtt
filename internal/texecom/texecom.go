package texecom

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/daemonp/texecom2mqtt/internal/log"
)

type Texecom struct {
	log           *log.Logger
	conn          net.Conn
	mu            sync.Mutex
	sequence      uint8
	eventChan     chan interface{}
	isConnected   bool
	disconnectChan chan struct{}
}

func NewTexecom(logger *log.Logger) *Texecom {
	return &Texecom{
		log:           logger,
		eventChan:     make(chan interface{}, 100),
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

	return nil
}

func (t *Texecom) GetPanelIdentification() (Device, error) {
	cmd := []byte{0x16} // Get Panel Identification command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return Device{}, fmt.Errorf("failed to get panel identification: %v", err)
	}

	// Parse the response to extract device information
	device := Device{
		Model:           string(resp[:20]),
		SerialNumber:    string(resp[20:40]),
		FirmwareVersion: string(resp[40:60]),
		Zones:           int(binary.LittleEndian.Uint16(resp[60:62])),
	}

	return device, nil
}

func (t *Texecom) GetAllAreas() ([]Area, error) {
	cmd := []byte{0x22} // Get Area Text command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get areas: %v", err)
	}

	var areas []Area
	for i := 0; i < len(resp); i += 16 {
		areaNumber := i/16 + 1
		areaName := string(resp[i : i+16])
		areas = append(areas, Area{
			Number: areaNumber,
			Name:   areaName,
			ID:     fmt.Sprintf("A%d", areaNumber),
		})
	}

	return areas, nil
}

func (t *Texecom) GetAllZones() ([]Zone, error) {
	cmd := []byte{0x03} // Get Zone Details command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %v", err)
	}

	var zones []Zone
	for i := 0; i < len(resp); i += 32 {
		zoneNumber := i/32 + 1
		zoneName := string(resp[i : i+16])
		zoneType := ZoneType(resp[i+16])
		zones = append(zones, Zone{
			Number: zoneNumber,
			Name:   zoneName,
			Type:   zoneType,
			ID:     fmt.Sprintf("Z%d", zoneNumber),
		})
	}

	return zones, nil
}

func (t *Texecom) GetZoneStates() ([]ZoneState, error) {
	cmd := []byte{0x02} // Get Zone State command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone states: %v", err)
	}

	var states []ZoneState
	for _, b := range resp {
		states = append(states, ZoneState(b&0x03))
	}

	return states, nil
}

func (t *Texecom) GetAreaStates() ([]AreaStatus, error) {
	cmd := []byte{0x0B} // Get Area Flags command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get area states: %v", err)
	}

	var states []AreaStatus
	for i := 0; i < len(resp); i += 8 {
		flags := binary.LittleEndian.Uint64(resp[i : i+8])
		status := AreaStatus{
			Status:  t.parseAreaState(flags),
			PartArm: t.parsePartArm(flags),
		}
		states = append(states, status)
	}

	return states, nil
}

func (t *Texecom) Arm(areaNumber int, armType ArmType) error {
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

func (t *Texecom) parseZoneEvent(data []byte) ZoneEvent {
	return ZoneEvent{
		ZoneNumber: int(binary.LittleEndian.Uint16(data[:2])),
		ZoneState:  ZoneState(data[2] & 0x03),
	}
}

func (t *Texecom) parseAreaEvent(data []byte) AreaEvent {
	return AreaEvent{
		AreaNumber: int(data[0]),
		AreaState:  AreaState(data[1]),
	}
}

func (t *Texecom) parseLogEvent(data []byte) LogEvent {
	return LogEvent{
		Type:        LogEventType(data[0]),
		GroupType:   LogEventGroupType(data[1]),
		Parameter:   binary.LittleEndian.Uint16(data[2:4]),
		Areas:       binary.LittleEndian.Uint16(data[4:6]),
		Time:        t.parseTimestamp(data[6:10]),
		Description: t.getLogEventDescription(LogEventType(data[0])),
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

func (t *Texecom) parseAreaState(flags uint64) AreaState {
	if flags&(1<<0) != 0 {
		return AreaStateInAlarm
	}
	if flags&(1<<21) != 0 || flags&(1<<22) != 0 || flags&(1<<23) != 0 {
		return AreaStateArmed
	}
	return AreaStateDisarmed
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

func (t *Texecom) getLogEventDescription(eventType LogEventType) string {
    switch eventType {
    case LogEventTypeEntryExit1:
        return "Entry/Exit 1"
    case LogEventTypeEntryExit2:
        return "Entry/Exit 2"
    case LogEventTypeGuard:
        return "Guard"
    case LogEventTypeGuardAccess:
        return "Guard Access"
    case LogEventTwentyFourHourAudible:
        return "24hr Audible"
    case LogEventTwentyFourHourSilent:
        return "24hr Silent"
    case LogEventPAAudible:
        return "PA Audible"
    case LogEventPASilent:
        return "PA Silent"
    case LogEventFire:
        return "Fire Alarm"
    case LogEventMedical:
        return "Medical"
    case LogEventTwentyFourHourGas:
        return "24Hr Gas Alarm"
    case LogEventAuxiliary:
        return "Auxiliary Alarm"
    case LogEventTamper:
        return "24hr Tamper Alarm"
    case LogEventExitTerminator:
        return "Exit Terminator"
    case LogEventMomentKey:
        return "Keyswitch - Momentary"
    case LogEventLatchKey:
        return "Keyswitch - Latching"
    case LogEventSecurity:
        return "Security Key"
    case LogEventOmitKey:
        return "Omit Key"
    case LogEventCustom:
        return "Custom Alarm"
    // Add more cases for other log event types
    default:
        return fmt.Sprintf("Unknown Log Event Type: %d", eventType)
    }
}

func (t *Texecom) getZoneTypeDescription(zoneType ZoneType) string {
    switch zoneType {
    case ZoneTypeNotUsed:
        return "Not used"
    case ZoneTypeEntryExit1:
        return "Entry/Exit 1"
    case ZoneTypeEntryExit2:
        return "Entry/Exit 2"
    case ZoneTypeGuard:
        return "Guard"
    case ZoneTypeGuardAccess:
        return "Guard Access"
    case ZoneTypeTwentyFourHourAudible:
        return "24Hr Audible"
    case ZoneTypeTwentyFourHourSilent:
        return "24Hr Silent"
    case ZoneTypePAAudible:
        return "PA Audible"
    case ZoneTypePASilent:
        return "PA Silent"
    case ZoneTypeFire:
        return "Fire"
    case ZoneTypeMedical:
        return "Medical"
    case ZoneTypeTwentyFourHourGas:
        return "24Hr Gas"
    case ZoneTypeAuxiliary:
        return "Auxiliary"
    case ZoneTypeTamper:
        return "Tamper"
    // Add more cases for other zone types
    default:
        return fmt.Sprintf("Unknown Zone Type: %d", zoneType)
    }
}

func (t *Texecom) getZoneStateDescription(zoneState ZoneState) string {
    switch zoneState {
    case ZoneStateSecure:
        return "Secure"
    case ZoneStateActive:
        return "Active"
    case ZoneStateTampered:
        return "Tampered"
    case ZoneStateShort:
        return "Short"
    default:
        return fmt.Sprintf("Unknown Zone State: %d", zoneState)
    }
}

func (t *Texecom) getAreaStateDescription(areaState AreaState) string {
    switch areaState {
    case AreaStateDisarmed:
        return "Disarmed"
    case AreaStateInExit:
        return "In Exit"
    case AreaStateInEntry:
        return "In Entry"
    case AreaStateArmed:
        return "Armed"
    case AreaStatePartArmed:
        return "Part Armed"
    case AreaStateInAlarm:
        return "In Alarm"
    default:
        return fmt.Sprintf("Unknown Area State: %d", areaState)
    }
}

func (t *Texecom) logZoneEvent(event ZoneEvent) {
    zoneState := t.getZoneStateDescription(event.ZoneState)
    t.log.Info("Zone %d state changed to %s", event.ZoneNumber, zoneState)
}

func (t *Texecom) logAreaEvent(event AreaEvent) {
    areaState := t.getAreaStateDescription(event.AreaState)
    t.log.Info("Area %d state changed to %s", event.AreaNumber, areaState)
}

func (t *Texecom) logLogEvent(event LogEvent) {
    t.log.Info("Log Event: %s (Type: %d, Group: %d, Parameter: %d, Areas: %d, Time: %s)",
        event.Description, event.Type, event.GroupType, event.Parameter, event.Areas, event.Time)
}
