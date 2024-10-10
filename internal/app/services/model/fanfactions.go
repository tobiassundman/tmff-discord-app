package model

// enum with strings "Off", "On", "On - no Fire & Ice".
type FanFactionSetting string

const (
	Off            FanFactionSetting = "Off"
	On             FanFactionSetting = "On"
	OnNoFireAndIce FanFactionSetting = "On - no Fire & Ice"
	Unknown        FanFactionSetting = "Unknown"
)

// Map to and from string.
var fanFactionSettingMap = map[string]FanFactionSetting{
	"Off":                  Off,
	"On":                   On,
	"On - with Fire & Ice": OnNoFireAndIce,
	"Unknown":              Unknown,
}

func FanFactionSettingFromString(s string) FanFactionSetting {
	value, ok := fanFactionSettingMap[s]
	if !ok {
		return Unknown
	}
	return value
}

func (f FanFactionSetting) String() string {
	return string(f)
}
