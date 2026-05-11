package app

import "testing"

func TestDefaultAlertsConfig(t *testing.T) {
	cfg := DefaultAlertsConfig()
	if cfg.CPUTempC != 85 {
		t.Errorf("CPUTempC default = %v, want 85", cfg.CPUTempC)
	}
	if cfg.GPUTempC != 85 {
		t.Errorf("GPUTempC default = %v, want 85", cfg.GPUTempC)
	}
	if cfg.PackagePowerW != 40 {
		t.Errorf("PackagePowerW default = %v, want 40", cfg.PackagePowerW)
	}
	if cfg.MemoryUsedPct != 90 {
		t.Errorf("MemoryUsedPct default = %v, want 90", cfg.MemoryUsedPct)
	}
	if cfg.HysteresisC != 3 {
		t.Errorf("HysteresisC default = %v, want 3", cfg.HysteresisC)
	}
	if cfg.HysteresisPct != 5 {
		t.Errorf("HysteresisPct default = %v, want 5", cfg.HysteresisPct)
	}
	if cfg.CooldownMs != 30000 {
		t.Errorf("CooldownMs default = %v, want 30000", cfg.CooldownMs)
	}
	if !cfg.Enabled {
		t.Errorf("Enabled default = false, want true")
	}
}
