package app

// AlertsConfig holds user-tunable alert thresholds. All values are absolute
// (Celsius for temps, Watts for power, percent for utilization/memory).
// Hysteresis fields prevent rapid Normal⇄Warning toggles around the boundary.
type AlertsConfig struct {
	Enabled       bool    `json:"enabled"`
	CPUTempC      float64 `json:"cpu_temp_c,omitempty"`
	GPUTempC      float64 `json:"gpu_temp_c,omitempty"`
	PackagePowerW float64 `json:"package_power_w,omitempty"`
	MemoryUsedPct float64 `json:"memory_used_pct,omitempty"`
	HysteresisC   float64 `json:"hysteresis_c,omitempty"`
	HysteresisPct float64 `json:"hysteresis_pct,omitempty"`
	CooldownMs    int     `json:"cooldown_ms,omitempty"`
}

func DefaultAlertsConfig() AlertsConfig {
	return AlertsConfig{
		Enabled:       true,
		CPUTempC:      85,
		GPUTempC:      85,
		PackagePowerW: 40,
		MemoryUsedPct: 90,
		HysteresisC:   3,
		HysteresisPct: 5,
		CooldownMs:    30000,
	}
}

// ResolveAlertsConfig returns defaults merged with non-zero user overrides.
// A nil override returns DefaultAlertsConfig() unchanged.
//
// Semantics for Enabled: a non-nil user config is taken at face value —
// if you write `"alerts": {"cpu_temp_c": 75}` and omit "enabled", you get
// Enabled=false. Document this in README: setting any alerts field means
// you must also set "enabled": true to keep alerts firing.
func ResolveAlertsConfig(user *AlertsConfig) AlertsConfig {
	out := DefaultAlertsConfig()
	if user == nil {
		return out
	}
	out.Enabled = user.Enabled
	if user.CPUTempC > 0 {
		out.CPUTempC = user.CPUTempC
	}
	if user.GPUTempC > 0 {
		out.GPUTempC = user.GPUTempC
	}
	if user.PackagePowerW > 0 {
		out.PackagePowerW = user.PackagePowerW
	}
	if user.MemoryUsedPct > 0 {
		out.MemoryUsedPct = user.MemoryUsedPct
	}
	if user.HysteresisC > 0 {
		out.HysteresisC = user.HysteresisC
	}
	if user.HysteresisPct > 0 {
		out.HysteresisPct = user.HysteresisPct
	}
	if user.CooldownMs > 0 {
		out.CooldownMs = user.CooldownMs
	}
	return out
}
