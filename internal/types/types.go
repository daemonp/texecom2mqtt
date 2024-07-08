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
	switch z {
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
	default:
		return fmt.Sprintf("Unknown ZoneType(%d)", z)
	}
}

type ZoneState int

const (
	ZoneStateSecure ZoneState = iota
	ZoneStateActive
	ZoneStateTampered
	ZoneStateShort
)

func (z ZoneState) String() string {
	switch z {
	case ZoneStateSecure:
		return "Secure"
	case ZoneStateActive:
		return "Active"
	case ZoneStateTampered:
		return "Tampered"
	case ZoneStateShort:
		return "Short"
	default:
		return fmt.Sprintf("Unknown ZoneState(%d)", z)
	}
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
	switch a {
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
		return fmt.Sprintf("Unknown AreaState(%d)", a)
	}
}

type ArmType int

const (
	ArmTypeFull ArmType = iota
	ArmTypePartArm1
	ArmTypePartArm2
	ArmTypePartArm3
)

func (a ArmType) String() string {
	switch a {
	case ArmTypeFull:
		return "Full Arm"
	case ArmTypePartArm1:
		return "Part Arm 1"
	case ArmTypePartArm2:
		return "Part Arm 2"
	case ArmTypePartArm3:
		return "Part Arm 3"
	default:
		return fmt.Sprintf("Unknown ArmType(%d)", a)
	}
}

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
