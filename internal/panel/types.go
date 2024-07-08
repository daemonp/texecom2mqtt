package panel

import "time"

// Device represents the Texecom device information.
type Device struct {
	Model           string
	SerialNumber    string
	FirmwareVersion string
	Zones           int
}

// Area represents an area in the Texecom system.
type Area struct {
	Number  int
	Name    string
	ID      string
	Status  AreaState
	PartArm int
}

// Zone represents a zone in the Texecom system.
type Zone struct {
	Number int
	Name   string
	Type   ZoneType
	ID     string
	Status ZoneState
}

// AreaStatus represents the status of an area.
type AreaStatus struct {
	Status  AreaState
	PartArm int
}

// ZoneEvent represents an event related to a zone.
type ZoneEvent struct {
	ZoneNumber int
	ZoneState  ZoneState
}

// AreaEvent represents an event related to an area.
type AreaEvent struct {
	AreaNumber int
	AreaState  AreaState
	PartArm    int
}

// LogEvent represents a log event in the Texecom system.
type LogEvent struct {
	Type        LogEventType
	GroupType   LogEventGroupType
	Parameter   uint16
	Areas       uint16
	Time        time.Time
	Description string
}

// ZoneType represents the type of a zone.
type ZoneType int

const (
	ZoneTypeNotUsed ZoneType = iota
	ZoneTypeEntryExit1
	ZoneTypeEntryExit2
	ZoneTypeGuard
	ZoneTypeGuardAccess
	ZoneTypeTwentyFourHourAudible
	ZoneTypeTwentyFourHourSilent
	ZoneTypePAAudible
	ZoneTypePASilent
	ZoneTypeFire
	ZoneTypeMedical
	ZoneTypeTwentyFourHourGas
	ZoneTypeAuxiliary
	ZoneTypeTamper
	// Add more zone types as needed
)

// ZoneState represents the state of a zone.
type ZoneState int

const (
	ZoneStateSecure ZoneState = iota
	ZoneStateActive
	ZoneStateTampered
	ZoneStateShort
)

// AreaState represents the state of an area.
type AreaState int

const (
	AreaStateDisarmed AreaState = iota
	AreaStateInExit
	AreaStateInEntry
	AreaStateArmed
	AreaStatePartArmed
	AreaStateInAlarm
)

// ArmType represents the type of arming.
type ArmType int

const (
	ArmTypeFull ArmType = iota
	ArmTypePartArm1
	ArmTypePartArm2
	ArmTypePartArm3
)

// LogEventType represents the type of a log event.
type LogEventType int

const (
	LogEventTypeEntryExit1 LogEventType = iota + 1
	LogEventTypeEntryExit2
	LogEventTypeGuard
	LogEventTypeGuardAccess
	LogEventTwentyFourHourAudible
	LogEventTwentyFourHourSilent
	LogEventPAAudible
	LogEventPASilent
	LogEventFire
	LogEventMedical
	LogEventTwentyFourHourGas
	LogEventAuxiliary
	LogEventTamper
	LogEventExitTerminator
	LogEventMomentKey
	LogEventLatchKey
	LogEventSecurity
	LogEventOmitKey
	LogEventCustom
	// Add more log event types as needed
)

// LogEventGroupType represents the group type of a log event.
type LogEventGroupType int

const (
	LogEventGroupTypeNotReported LogEventGroupType = iota
	LogEventGroupTypePriorityAlarm
	LogEventGroupTypePriorityAlarmRestore
	LogEventGroupTypeAlarm
	LogEventGroupTypeRestore
	LogEventGroupTypeOpen
	LogEventGroupTypeClose
	// Add more log event group types as needed
)
