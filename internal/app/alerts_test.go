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

func TestAppConfigCarriesAlerts(t *testing.T) {
	cfg := AppConfig{Alerts: &AlertsConfig{CPUTempC: 70, Enabled: true}}
	if cfg.Alerts == nil {
		t.Fatal("Alerts nil")
	}
	if cfg.Alerts.CPUTempC != 70 {
		t.Errorf("CPUTempC = %v, want 70", cfg.Alerts.CPUTempC)
	}
}

func TestResolveAlertsConfigFallsBackToDefaults(t *testing.T) {
	got := ResolveAlertsConfig(nil)
	want := DefaultAlertsConfig()
	if got != want {
		t.Errorf("ResolveAlertsConfig(nil) = %+v, want %+v", got, want)
	}
}

func TestResolveAlertsConfigMergesUserOverrides(t *testing.T) {
	override := &AlertsConfig{Enabled: true, CPUTempC: 75}
	got := ResolveAlertsConfig(override)
	if got.CPUTempC != 75 {
		t.Errorf("CPUTempC = %v, want 75", got.CPUTempC)
	}
	if got.GPUTempC != 85 {
		t.Errorf("GPUTempC = %v, want 85 (default kept)", got.GPUTempC)
	}
	if !got.Enabled {
		t.Errorf("Enabled = false, want true (override said true)")
	}
}

func TestResolveAlertsConfigUserCanDisable(t *testing.T) {
	override := &AlertsConfig{Enabled: false, CPUTempC: 75}
	got := ResolveAlertsConfig(override)
	if got.Enabled {
		t.Errorf("Enabled = true, want false (override said false)")
	}
}
