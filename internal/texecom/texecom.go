package texecom

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/daemonp/texecom2mqtt/internal/log"
	"github.com/daemonp/texecom2mqtt/internal/types"
)

type Texecom struct {
	// conn net.Conn
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

const (
	CMD_TIMEOUT = 50000 * time.Millisecond
	CMD_RETRIES = 5
)

func (t *Texecom) Connect(host string, port int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.log.Debug("Attempting to connect to %s:%d", host, port)
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		t.log.Error("Connection failed: %v", err)
		return fmt.Errorf("failed to connect: %v", err)
	}

	t.conn = conn
	t.isConnected = true // Set this flag
	t.log.Debug("Connection established")

	// Get the serial number
	serialNumber, err := t.getSerialNumber(ctx)
	if err != nil {
		t.log.Error("Failed to get serial number: %v", err)
		t.Disconnect()
		return fmt.Errorf("failed to get serial number: %v", err)
	}
	t.log.Info("Retrieved serial number: %s", serialNumber)

	// Check connection status after getting serial number
	if !t.isConnected {
		return fmt.Errorf("connection lost after retrieving serial number")
	}

	go t.readLoop()

	return nil
}

func (t *Texecom) Disconnect() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return
	}

	if t.conn != nil {
		t.conn.Close()
	}

	t.log.Debug("Disconnecting from panel")
	close(t.disconnectChan)
	t.conn.Close()
	t.isConnected = false
	close(t.eventChan)
	t.log.Debug("Disconnected from panel")
}

func (t *Texecom) Login(password string) error {
	if !t.isConnected {
		return fmt.Errorf("not connected to panel")
	}

	t.log.Debug("Preparing login command")
	cmd := []byte{0x01} // Login command
	cmd = append(cmd, []byte(password)...)

	packet := t.createCommandPacket(cmd[0], cmd[1:])
	t.log.Debug("Login packet: %x", packet)

	response, err := t.sendCommand(packet)
	if err != nil {
		t.log.Error("Failed to send login command: %v", err)
		return fmt.Errorf("failed to send login command: %v", err)
	}

	t.log.Debug("Received login response: %x", response)
	if len(response) > 0 && response[5] == 0x06 { // ACK
		t.isLoggedIn = true
		t.log.Info("Login successful")
		return nil
	}
	return fmt.Errorf("login failed: invalid response")
}

// func (t *Texecom) sendCommand(packet []byte) ([]byte, error) {
//     t.mu.Lock()
//     defer t.mu.Unlock()

//     if !t.isConnected {
//         return nil, fmt.Errorf("not connected")
//     }

//     t.log.Debug("Sending command: %x", packet)
//     _, err := t.conn.Write(packet)
//     if err != nil {
//         t.log.Error("Failed to send command: %v", err)
//         t.isConnected = false
//         return nil, fmt.Errorf("failed to send command: %v", err)
//     }

//     t.log.Debug("Waiting for response")
//     deadline := time.Now().Add(10 * time.Second)
//     for time.Now().Before(deadline) {
//         t.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
//         resp := make([]byte, 1024)
//         n, err := t.conn.Read(resp)
//         if err != nil {
//             if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
//                 continue // This is just a timeout for this read attempt, not for the whole operation
//             }
//             t.log.Error("Failed to read response: %v", err)
//             t.isConnected = false
//             return nil, fmt.Errorf("failed to read response: %v", err)
//         }

//         t.log.Debug("Received response: %x", resp[:n])
//         // Check if this is the response we're looking for
//         if n >= 3 && resp[0] == 't' && resp[1] == 'R' {
//             return resp[:n], nil
//         }
//         // If it's not the response we're looking for, process it and continue waiting
//         t.processMessage(resp[:n])
//     }

//     return nil, fmt.Errorf("command timed out")
// }

func (t *Texecom) sendCommand(packet []byte) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return nil, fmt.Errorf("not connected")
	}

	t.log.Debug("Sending command: %x", packet)
	_, err := t.conn.Write(packet)
	if err != nil {
		t.log.Error("Failed to send command: %v", err)
		t.isConnected = false
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	t.log.Debug("Waiting for response")
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		t.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		resp := make([]byte, 1024)
		n, err := t.conn.Read(resp)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // This is just a timeout for this read attempt, not for the whole operation
			}
			t.log.Error("Failed to read response: %v", err)
			t.isConnected = false
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		t.log.Debug("Received response: %x", resp[:n])
		if n >= 3 && resp[0] == 't' && resp[1] == 'R' {
			return resp[:n], nil
		}
		t.processMessage(resp[:n])
	}

	return nil, fmt.Errorf("command timed out")
}

func (t *Texecom) sendCommandAndWaitForResponse(packet []byte, timeout time.Duration) ([]byte, error) {
	err := t.sendRawCommand(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	deadline := time.Now().Add(timeout)
	buffer := make([]byte, 1024)

	for time.Now().Before(deadline) {
		t.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := t.conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // This is just a timeout for this read attempt, not for the whole operation
			}
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		t.log.Debug("Received data: %v", buffer[:n])
		return buffer[:n], nil
	}

	return nil, fmt.Errorf("response not received within timeout")
}

func (t *Texecom) sendCommandWithTimeout(packet []byte, timeout time.Duration) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return nil, fmt.Errorf("not connected")
	}

	t.log.Debug("Sending command: %v", packet)
	_, err := t.conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	t.conn.SetReadDeadline(time.Now().Add(timeout))
	defer t.conn.SetReadDeadline(time.Time{}) // Reset the deadline

	resp := make([]byte, 1024)
	n, err := t.conn.Read(resp)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("timeout while reading response")
		}
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return resp[:n], nil
}

func (t *Texecom) GetPanelIdentification() (types.Device, error) {
	t.log.Debug("Sending Get Panel Identification command")
	cmd := []byte{0x16} // Get Panel Identification command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to get panel identification: %v", err)
		return types.Device{}, fmt.Errorf("failed to get panel identification: %v", err)
	}

	t.log.Debug("Parsing panel identification response")
	device := types.Device{
		Model:           string(resp[:20]),
		SerialNumber:    string(resp[20:40]),
		FirmwareVersion: string(resp[40:60]),
		Zones:           int(binary.LittleEndian.Uint16(resp[60:62])),
	}

	t.device = device
	t.log.Debug("Panel identification: %+v", device)
	return device, nil
}

func (t *Texecom) GetAllAreas() ([]types.Area, error) {
	t.log.Debug("Sending Get Area Text command")
	cmd := []byte{0x22} // Get Area Text command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to get areas: %v", err)
		return nil, fmt.Errorf("failed to get areas: %v", err)
	}

	t.log.Debug("Parsing area information")
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
	t.log.Debug("Retrieved %d areas", len(areas))
	return areas, nil
}

func (t *Texecom) GetAllZones() ([]types.Zone, error) {
	t.log.Debug("Sending Get Zone Details command")
	cmd := []byte{0x03} // Get Zone Details command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to get zones: %v", err)
		return nil, fmt.Errorf("failed to get zones: %v", err)
	}

	t.log.Debug("Parsing zone information")
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
	t.log.Debug("Retrieved %d zones", len(zones))
	return zones, nil
}

func (t *Texecom) GetZoneStates() ([]types.ZoneState, error) {
	t.log.Debug("Sending Get Zone State command")
	cmd := []byte{0x02} // Get Zone State command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to get zone states: %v", err)
		return nil, fmt.Errorf("failed to get zone states: %v", err)
	}

	t.log.Debug("Parsing zone states")
	var states []types.ZoneState
	for _, b := range resp {
		states = append(states, types.ZoneState(b&0x03))
	}

	t.log.Debug("Retrieved states for %d zones", len(states))
	return states, nil
}

func (t *Texecom) GetAreaStates() ([]types.AreaStatus, error) {
	t.log.Debug("Sending Get Area Flags command")
	cmd := []byte{0x0B} // Get Area Flags command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to get area states: %v", err)
		return nil, fmt.Errorf("failed to get area states: %v", err)
	}

	t.log.Debug("Parsing area states")
	var states []types.AreaStatus
	for i := 0; i < len(resp); i += 8 {
		flags := binary.LittleEndian.Uint64(resp[i : i+8])
		status := types.AreaStatus{
			Status:  t.parseAreaState(flags),
			PartArm: t.parsePartArm(flags),
		}
		states = append(states, status)
	}

	t.log.Debug("Retrieved states for %d areas", len(states))
	return states, nil
}

func (t *Texecom) Arm(areaNumber int, armType types.ArmType) error {
	t.log.Debug("Sending Arm Area command for area %d, type %v", areaNumber, armType)
	cmd := []byte{0x06, byte(areaNumber), byte(armType)} // Arm Area command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to arm area: %v", err)
		return fmt.Errorf("failed to arm area: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		t.log.Error("Failed to arm area: invalid response")
		return fmt.Errorf("failed to arm area: invalid response")
	}

	t.log.Debug("Area armed successfully")
	return nil
}

func (t *Texecom) Disarm(areaNumber int) error {
	t.log.Debug("Sending Disarm Area command for area %d", areaNumber)
	cmd := []byte{0x08, byte(areaNumber)} // Disarm Area command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to disarm area: %v", err)
		return fmt.Errorf("failed to disarm area: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		t.log.Error("Failed to disarm area: invalid response")
		return fmt.Errorf("failed to disarm area: invalid response")
	}

	t.log.Debug("Area disarmed successfully")
	return nil
}

func (t *Texecom) Reset(areaNumber int) error {
	t.log.Debug("Sending Reset Area command for area %d", areaNumber)
	cmd := []byte{0x09, byte(areaNumber)} // Reset Area command
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to reset area: %v", err)
		return fmt.Errorf("failed to reset area: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		t.log.Error("Failed to reset area: invalid response")
		return fmt.Errorf("failed to reset area: invalid response")
	}

	t.log.Debug("Area reset successfully")
	return nil
}

func (t *Texecom) SetDateTime(datetime time.Time) error {
	t.log.Debug("Sending Set Date/Time command for %v", datetime)
	cmd := []byte{0x18} // Set Date/Time command
	cmd = append(cmd, byte(datetime.Day()), byte(datetime.Month()), byte(datetime.Year()%100),
		byte(datetime.Hour()), byte(datetime.Minute()), byte(datetime.Second()))
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to set date/time: %v", err)
		return fmt.Errorf("failed to set date/time: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		t.log.Error("Failed to set date/time: invalid response")
		return fmt.Errorf("failed to set date/time: invalid response")
	}

	t.log.Debug("Date/Time set successfully")
	return nil
}

func (t *Texecom) SetLCDDisplay(text string) error {
	t.log.Debug("Sending Set LCD Display command with text: %s", text)
	if len(text) > 32 {
		text = text[:32]
	}
	cmd := []byte{0x0E} // Set LCD Display command
	cmd = append(cmd, []byte(text)...)
	cmd = append(cmd, make([]byte, 32-len(text))...) // Pad with spaces
	resp, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to set LCD display: %v", err)
		return fmt.Errorf("failed to set LCD display: %v", err)
	}

	if resp[0] != 0x06 { // ACK
		t.log.Error("Failed to set LCD display: invalid response")
		return fmt.Errorf("failed to set LCD display: invalid response")
	}

	t.log.Debug("LCD Display set successfully")
	return nil
}

func (t *Texecom) UpdateSystemPower() error {
	t.log.Debug("Sending Get System Power command")
	cmd := []byte{0x19} // Get System Power command
	_, err := t.sendCommand(cmd)
	if err != nil {
		t.log.Error("Failed to update system power: %v", err)
		return fmt.Errorf("failed to update system power: %v", err)
	}

	t.log.Debug("System power updated successfully")
	return nil
}

func (t *Texecom) Events() <-chan interface{} {
	return t.eventChan
}

func (t *Texecom) readLoop() {
	buffer := make([]byte, 1024)
	for {
		select {
		case <-t.disconnectChan:
			return
		default:
			t.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, err := t.conn.Read(buffer)
			t.conn.SetReadDeadline(time.Time{}) // Reset the deadline
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is a timeout, which is expected when there's no data
					continue
				}
				t.log.Error("Read error: %v", err)
				t.Disconnect()
				return
			}

			t.processMessage(buffer[:n])
		}
	}
}

// func (t *Texecom) processMessage(msg []byte) {
//    t.log.Debug("Processing message: %v", msg)
//  	if len(msg) < 5 {
// 		return
// 	}

// 	t.log.Debug("Processing message: %v", msg)

// 	switch msg[1] {
// 	case 'M': // Event message
// 		event := t.parseEvent(msg[4:])
// 		t.eventChan <- event
// 	case 'R': // Response message
// 		// Handle responses if needed
// 		t.log.Debug("Received response message: %v", msg)
// 	}
// }

func (t *Texecom) validateCrc(message []byte) bool {
	crc := message[len(message)-1]
	return t.calculateCRC(message[:len(message)-1]) == crc
}

func (t *Texecom) processMessage(msg []byte) {
	t.log.Debug("Processing message: %v", msg)
	if len(msg) < 5 {
		return
	}

	if !t.validateCrc(msg) {
		t.log.Error("Invalid CRC for message: %x", msg)
		return
	}

	switch msg[1] {
	case 'M': // Event message
		event := t.parseEvent(msg[4:])
		t.eventChan <- event
	case 'R': // Response message
		t.log.Debug("Received response message: %v", msg)
	}
}

func (t *Texecom) parseEvent(data []byte) interface{} {
	if len(data) < 2 {
		t.log.Warn("Received event data is too short")
		return nil
	}

	eventType := data[0]
	t.log.Debug("Parsing event of type: %d", eventType)

	switch eventType {
	case 1: // Zone event
		return t.parseZoneEvent(data[1:])
	case 2: // Area event
		return t.parseAreaEvent(data[1:])
	case 5: // Log event
		return t.parseLogEvent(data[1:])
	default:
		t.log.Warn("Unknown event type: %d", eventType)
		return nil
	}
}

func (t *Texecom) parseZoneEvent(data []byte) types.ZoneEvent {
	event := types.ZoneEvent{
		ZoneNumber: int(binary.LittleEndian.Uint16(data[:2])),
		ZoneState:  types.ZoneState(data[2] & 0x03),
	}
	t.log.Debug("Parsed Zone Event: %+v", event)
	return event
}

func (t *Texecom) parseAreaEvent(data []byte) types.AreaEvent {
	event := types.AreaEvent{
		AreaNumber: int(data[0]),
		AreaState:  types.AreaState(data[1]),
	}
	t.log.Debug("Parsed Area Event: %+v", event)
	return event
}

func (t *Texecom) parseLogEvent(data []byte) types.LogEvent {
	event := types.LogEvent{
		Type:        types.LogEventType(data[0]),
		GroupType:   types.LogEventGroupType(data[1]),
		Parameter:   binary.LittleEndian.Uint16(data[2:4]),
		Areas:       binary.LittleEndian.Uint16(data[4:6]),
		Time:        t.parseTimestamp(data[6:10]),
		Description: t.getLogEventDescription(types.LogEventType(data[0])),
	}
	t.log.Debug("Parsed Log Event: %+v", event)
	return event
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
	crc := byte(0xFF)
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
	description := LogEventTypeDescriptions[eventType]
	if description == "" {
		description = fmt.Sprintf("Unknown Log Event Type: %d", eventType)
	}
	return description
}

func (t *Texecom) createCommandPacket(command byte, body []byte) []byte {
	length := byte(6 + len(body))
	packet := make([]byte, length)
	packet[0] = 't' // HEADER_START
	packet[1] = 'C' // HeaderType.Command
	packet[2] = length
	packet[3] = t.sequence
	packet[4] = command
	if len(body) > 0 {
		copy(packet[5:], body)
	}
	crc := t.calculateCRC(packet[:length-1])
	packet[length-1] = crc
	t.sequence++
	return packet
}

func (t *Texecom) decodeSerialNumber(data []byte) string {
	if len(data) < 11 {
		return "Unknown"
	}
	// The serial number is in the last 7 bytes of the response
	serialBytes := data[4:11]
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x",
		serialBytes[0], serialBytes[1], serialBytes[2], serialBytes[3],
		serialBytes[4], serialBytes[5], serialBytes[6])
}

func (t *Texecom) getSerialNumber(ctx context.Context) (string, error) {
	t.log.Debug("Preparing to execute serial number command")
	payload := []byte{0x03, 0x5a, 0xa2}

	time.Sleep(1 * time.Second)

	t.log.Debug("Sending serial number command (raw)")
	err := t.sendRawCommand(payload)
	if err != nil {
		return "", fmt.Errorf("failed to send serial number command: %v", err)
	}

	t.log.Debug("Waiting for serial number response")

	responseChan := make(chan []byte)
	errorChan := make(chan error)

	go func() {
		buffer := make([]byte, 1024)
		n, err := t.conn.Read(buffer)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- buffer[:n]
	}()

	select {
	case response := <-responseChan:
		t.log.Debug("Received data: %x", response)
		if len(response) >= 4 && response[0] == 0x0b && response[1] == 0x5a {
			serialNumber := t.decodeSerialNumber(response)
			t.log.Debug("Parsed serial number: %s", serialNumber)
			return serialNumber, nil
		}
		return "", fmt.Errorf("unexpected response: %x", response)
	case err := <-errorChan:
		return "", fmt.Errorf("error reading response: %v", err)
	case <-ctx.Done():
		return "", fmt.Errorf("operation timed out")
	}
}

func (t *Texecom) sendRawCommand(payload []byte) error {
	t.log.Debug("Sending raw command: %x", payload)
	n, err := t.conn.Write(payload)
	if err != nil {
		t.log.Error("Failed to send raw command: %v", err)
		return fmt.Errorf("failed to send raw command: %v", err)
	}
	t.log.Debug("Sent %d bytes", n)
	return nil
}
