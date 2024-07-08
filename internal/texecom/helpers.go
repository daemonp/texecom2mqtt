package texecom

import (
	"encoding/binary"
	"time"
)

// ParseZoneBitmap parses a zone bitmap and returns a struct with the parsed data
func ParseZoneBitmap(zoneBitmap byte) ZoneBitmapData {
	return ZoneBitmapData{
		State:          ZoneState(zoneBitmap & 0x3),
		Fault:          (zoneBitmap & (1 << 2)) != 0,
		FailedTest:     (zoneBitmap & (1 << 3)) != 0,
		Alarmed:        (zoneBitmap & (1 << 4)) != 0,
		ManualBypassed: (zoneBitmap & (1 << 5)) != 0,
		AutoBypassed:   (zoneBitmap & (1 << 6)) != 0,
		Masked:         (zoneBitmap & (1 << 7)) != 0,
	}
}

// CalculateAreaSize calculates the size of the area bitmap based on the number of zones
func CalculateAreaSize(numberOfZones int) int {
	return (numberOfZones + 7) / 8
}

// CalculateZoneNumberSize calculates the size of the zone number field based on the number of zones
func CalculateZoneNumberSize(numberOfZones int) int {
	if numberOfZones > 256 {
		return 2
	}
	return 1
}

// CreateArmInput creates the input for the Arm command
func CreateArmInput(numberOfZones, area int, armType ArmType) []byte {
	size := CalculateAreaSize(numberOfZones)
	buffer := make([]byte, size+1)
	buffer[0] = byte(armType)
	WriteAreaBitmapToBuffer(size, area, buffer, 1)
	return buffer
}

// CreateDisarmOrResetInput creates the input for the Disarm or Reset command
func CreateDisarmOrResetInput(numberOfZones, area int) []byte {
	size := CalculateAreaSize(numberOfZones)
	buffer := make([]byte, size)
	WriteAreaBitmapToBuffer(size, area, buffer, 0)
	return buffer
}

// WriteAreaBitmapToBuffer writes the area bitmap to the given buffer
func WriteAreaBitmapToBuffer(size, area int, buffer []byte, offset int) {
	if size == 8 {
		bitmap := uint64(1) << uint(area)
		binary.LittleEndian.PutUint64(buffer[offset:], bitmap)
	} else {
		bitmap := uint32(1) << uint(area)
		binary.LittleEndian.PutUint32(buffer[offset:], bitmap)
	}
}

// ParseTimestamp parses a Texecom timestamp and returns a time.Time
func ParseTimestamp(data []byte) time.Time {
	timestamp := binary.LittleEndian.Uint32(data)
	seconds := timestamp & 63
	minutes := (timestamp >> 6) & 63
	hours := (timestamp >> 12) & 31
	day := (timestamp >> 17) & 31
	month := (timestamp >> 22) & 15
	year := 2000 + ((timestamp >> 26) & 63)

	return time.Date(int(year), time.Month(month), int(day), int(hours), int(minutes), int(seconds), 0, time.UTC)
}

// CreateSetDateInput creates the input for the SetDate command
func CreateSetDateInput(date time.Time) []byte {
	return []byte{
		byte(date.Day()),
		byte(date.Month()),
		byte(date.Year() % 100),
		byte(date.Hour()),
		byte(date.Minute()),
		byte(date.Second()),
	}
}

// CreateSetLCDDisplayInput creates the input for the SetLCDDisplay command
func CreateSetLCDDisplayInput(text string) []byte {
	if len(text) > 32 {
		text = text[:32]
	}
	input := make([]byte, 32)
	copy(input, text)
	return input
}

