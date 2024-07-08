package texecom

import "fmt"

// ArmTypeDescriptions maps ArmType to its string description
var ArmTypeDescriptions = map[ArmType]string{
	ArmTypeFull:     "Full Arm",
	ArmTypePartArm1: "Part Arm 1",
	ArmTypePartArm2: "Part Arm 2",
	ArmTypePartArm3: "Part Arm 3",
}

// AreaStateDescriptions maps AreaState to its string description
var AreaStateDescriptions = map[AreaState]string{
	AreaStateDisarmed:  "Disarmed",
	AreaStateInExit:    "In Exit",
	AreaStateInEntry:   "In Entry",
	AreaStateArmed:     "Armed",
	AreaStatePartArmed: "Part Armed",
	AreaStateInAlarm:   "In Alarm",
}

// ZoneStateDescriptions maps ZoneState to its string description
var ZoneStateDescriptions = map[ZoneState]string{
	ZoneStateSecure:   "Secure",
	ZoneStateActive:   "Active",
	ZoneStateTampered: "Tampered",
	ZoneStateShort:    "Short",
}

// ZoneTypeDescriptions maps ZoneType to its string description
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

// LogEventTypeDescriptions maps LogEventType to its string description
var LogEventTypeDescriptions = map[LogEventType]string{
	LogEventTypeEntryExit1:        "Entry/Exit 1",
	LogEventTypeEntryExit2:        "Entry/Exit 2",
	LogEventTypeGuard:             "Guard",
	LogEventTypeGuardAccess:       "Guard Access",
	LogEventTwentyFourHourAudible: "24hr Audible",
	LogEventTwentyFourHourSilent:  "24hr Silent",
	LogEventPAAudible:             "PA Audible",
	LogEventPASilent:              "PA Silent",
	LogEventFire:                  "Fire Alarm",
	LogEventMedical:               "Medical",
	LogEventTwentyFourHourGas:     "24Hr Gas Alarm",
	LogEventAuxiliary:             "Auxiliary Alarm",
	LogEventTamper:                "24hr Tamper Alarm",
	LogEventExitTerminator:        "Exit Terminator",
	LogEventMomentKey:             "Keyswitch - Momentary",
	LogEventLatchKey:              "Keyswitch - Latching",
	LogEventSecurity:              "Security Key",
	LogEventOmitKey:               "Omit Key",
	LogEventCustom:                "Custom Alarm",
	// Add more cases for other log event types
}

// GetAreaStatus returns a string description of the area status
func GetAreaStatus(area Area) string {
	status := AreaStateDescriptions[area.Status]
	if area.Status == AreaStatePartArmed {
		return fmt.Sprintf("%s %d", status, area.PartArm)
	}
	return status
}
