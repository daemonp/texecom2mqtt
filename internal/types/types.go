package types

import (
	"fmt"
	"time"
)

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
	Number        int
	Name          string
	Type          ZoneType
	ID            string
	Status        ZoneState
	HomeAssistant *HomeAssistantZone
}

type HomeAssistantZone struct {
	DeviceClass string `yaml:"device_class"`
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

func (t ZoneType) String() string {
	return ZoneTypeDescriptions[t]
}

func (s AreaState) String() string {
	return AreaStateDescriptions[s]
}

func (s ZoneState) String() string {
	return ZoneStateDescriptions[s]
}

var ArmTypeDescriptions = map[ArmType]string{
	ArmTypeFull:     "Full Arm",
	ArmTypePartArm1: "Part Arm 1",
	ArmTypePartArm2: "Part Arm 2",
	ArmTypePartArm3: "Part Arm 3",
}

var AreaStateDescriptions = map[AreaState]string{
	AreaStateDisarmed:  "Disarmed",
	AreaStateInExit:    "In Exit",
	AreaStateInEntry:   "In Entry",
	AreaStateArmed:     "Armed",
	AreaStatePartArmed: "Part Armed",
	AreaStateInAlarm:   "In Alarm",
}

var ZoneStateDescriptions = map[ZoneState]string{
	ZoneStateSecure:   "Secure",
	ZoneStateActive:   "Active",
	ZoneStateTampered: "Tampered",
	ZoneStateShort:    "Short",
}

var ZoneTypeDescriptions = map[ZoneType]string{
	ZoneTypeNotUsed:               "Not used",
	ZoneTypeEntryExit1:            "Entry/Exit 1",
	ZoneTypeEntryExit2:            "Entry/Exit 2",
	ZoneTypeGuard:                 "Guard",
	ZoneTypeGuardAccess:           "Guard Access",
	ZoneTypeTwentyFourHourAudible: "24Hr Audible",
	ZoneTypeTwentyFourHourSilent:  "24Hr Silent",
	ZoneTypePAAudible:             "PA Audible",
	ZoneTypePASilent:              "PA Silent",
	ZoneTypeFire:                  "Fire",
	ZoneTypeMedical:               "Medical",
	ZoneTypeTwentyFourHourGas:     "24Hr Gas",
	ZoneTypeAuxiliary:             "Auxiliary",
	ZoneTypeTamper:                "Tamper",
}

func GetAreaStatus(area Area) string {
	status := AreaStateDescriptions[area.Status]
	if area.Status == AreaStatePartArmed {
		return fmt.Sprintf("%s %d", status, area.PartArm)
	}
	return status
}
