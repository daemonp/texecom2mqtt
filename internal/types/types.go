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

func (z ZoneType) String() string {
	return [...]string{
		"Not used",
		"Entry/Exit 1",
		"Entry/Exit 2",
		"Guard",
		"Guard Access",
		"24Hr Audible",
		"24Hr Silent",
		"PA Audible",
		"PA Silent",
		"Fire",
		"Medical",
		"24Hr Gas",
		"Auxiliary",
		"Tamper",
	}[z]
}

type ZoneState int

const (
	ZoneStateSecure ZoneState = iota
	ZoneStateActive
	ZoneStateTampered
	ZoneStateShort
)

func (z ZoneState) String() string {
	return [...]string{
		"Secure",
		"Active",
		"Tampered",
		"Short",
	}[z]
}

type AreaState int

const (
	AreaStateDisarmed AreaState = iota
	AreaStateInExit
	AreaStateInEntry
	AreaStateArmed
	AreaStatePartArmed
	AreaStateInAlarm
)

func (a AreaState) String() string {
	return [...]string{
		"Disarmed",
		"In Exit",
		"In Entry",
		"Armed",
		"Part Armed",
		"In Alarm",
	}[a]
}

type ArmType int

const (
	ArmTypeFull ArmType = iota
	ArmTypePartArm1
	ArmTypePartArm2
	ArmTypePartArm3
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

type CacheData struct {
	Device     Device    `json:"device"`
	Areas      []Area    `json:"areas"`
	Zones      []Zone    `json:"zones"`
	LastUpdate time.Time `json:"last_update"`
}

