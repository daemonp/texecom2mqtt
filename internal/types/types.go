package types

import "time"

type Device struct {
	Model           string
	SerialNumber    string
	FirmwareVersion string
	Zones           int
}

type Area struct {
	Number  int
	Name    string
	ID      string
	Status  AreaState
	PartArm int
}

type Zone struct {
	Number int
	Name   string
	Type   ZoneType
	ID     string
	Status ZoneState
}

type AreaStatus struct {
	Status  AreaState
	PartArm int
}

type ZoneEvent struct {
	ZoneNumber int
	ZoneState  ZoneState
}

type AreaEvent struct {
	AreaNumber int
	AreaState  AreaState
	PartArm    int
}

type LogEvent struct {
	Type        LogEventType
	GroupType   LogEventGroupType
	Parameter   uint16
	Areas       uint16
	Time        time.Time
	Description string
}

type CacheData struct {
	Device     Device    `json:"device"`
	Areas      []Area    `json:"areas"`
	Zones      []Zone    `json:"zones"`
	LastUpdate time.Time `json:"last_update"`
}

type ArmType int

const (
	ArmTypeFull ArmType = iota
	ArmTypePartArm1
	ArmTypePartArm2
	ArmTypePartArm3
)

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
)

type ZoneState int

const (
	ZoneStateSecure ZoneState = iota
	ZoneStateActive
	ZoneStateTampered
	ZoneStateShort
)

type AreaState int

const (
	AreaStateDisarmed AreaState = iota
	AreaStateInExit
	AreaStateInEntry
	AreaStateArmed
	AreaStatePartArmed
	AreaStateInAlarm
)

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
)

type LogEventGroupType int

const (
	LogEventGroupTypeNotReported LogEventGroupType = iota
	LogEventGroupTypePriorityAlarm
	LogEventGroupTypePriorityAlarmRestore
	LogEventGroupTypeAlarm
	LogEventGroupTypeRestore
	LogEventGroupTypeOpen
	LogEventGroupTypeClose
)
