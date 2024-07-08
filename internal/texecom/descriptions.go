package texecom

import (
	"fmt"
	"github.com/daemonp/texecom2mqtt/internal/types"
)

var ArmTypeDescriptions = map[types.ArmType]string{
	types.ArmTypeFull:     "Full Arm",
	types.ArmTypePartArm1: "Part Arm 1",
	types.ArmTypePartArm2: "Part Arm 2",
	types.ArmTypePartArm3: "Part Arm 3",
}

var AreaStateDescriptions = map[types.AreaState]string{
	types.AreaStateDisarmed:  "Disarmed",
	types.AreaStateInExit:    "In Exit",
	types.AreaStateInEntry:   "In Entry",
	types.AreaStateArmed:     "Armed",
	types.AreaStatePartArmed: "Part Armed",
	types.AreaStateInAlarm:   "In Alarm",
}

var ZoneStateDescriptions = map[types.ZoneState]string{
	types.ZoneStateSecure:   "Secure",
	types.ZoneStateActive:   "Active",
	types.ZoneStateTampered: "Tampered",
	types.ZoneStateShort:    "Short",
}

var ZoneTypeDescriptions = map[types.ZoneType]string{
	types.ZoneTypeNotUsed:               "Not used",
	types.ZoneTypeEntryExit1:            "Entry/Exit 1",
	types.ZoneTypeEntryExit2:            "Entry/Exit 2",
	types.ZoneTypeGuard:                 "Guard",
	types.ZoneTypeGuardAccess:           "Guard Access",
	types.ZoneTypeTwentyFourHourAudible: "24Hr Audible",
	types.ZoneTypeTwentyFourHourSilent:  "24Hr Silent",
	types.ZoneTypePAAudible:             "PA Audible",
	types.ZoneTypePASilent:              "PA Silent",
	types.ZoneTypeFire:                  "Fire",
	types.ZoneTypeMedical:               "Medical",
	types.ZoneTypeTwentyFourHourGas:     "24Hr Gas",
	types.ZoneTypeAuxiliary:             "Auxiliary",
	types.ZoneTypeTamper:                "Tamper",
}

var LogEventTypeDescriptions = map[types.LogEventType]string{
	types.LogEventTypeEntryExit1:        "Entry/Exit 1",
	types.LogEventTypeEntryExit2:        "Entry/Exit 2",
	types.LogEventTypeGuard:             "Guard",
	types.LogEventTypeGuardAccess:       "Guard Access",
	types.LogEventTwentyFourHourAudible: "24hr Audible",
	types.LogEventTwentyFourHourSilent:  "24hr Silent",
	types.LogEventPAAudible:             "PA Audible",
	types.LogEventPASilent:              "PA Silent",
	types.LogEventFire:                  "Fire Alarm",
	types.LogEventMedical:               "Medical",
	types.LogEventTwentyFourHourGas:     "24Hr Gas Alarm",
	types.LogEventAuxiliary:             "Auxiliary Alarm",
	types.LogEventTamper:                "24hr Tamper Alarm",
	types.LogEventExitTerminator:        "Exit Terminator",
	types.LogEventMomentKey:             "Keyswitch - Momentary",
	types.LogEventLatchKey:              "Keyswitch - Latching",
	types.LogEventSecurity:              "Security Key",
	types.LogEventOmitKey:               "Omit Key",
	types.LogEventCustom:                "Custom Alarm",
	// Add more cases for other log event types
}

func GetAreaStatus(area types.Area) string {
	status := AreaStateDescriptions[area.Status]
	if area.Status == types.AreaStatePartArmed {
		return fmt.Sprintf("%s %d", status, area.PartArm)
	}
	return status
}
