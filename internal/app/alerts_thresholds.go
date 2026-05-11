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
