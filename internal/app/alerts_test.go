package app

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/metaspartan/mactop/v2/internal/i18n"
)

func init() { i18n.Init("en") }

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

func TestAlertEventStringer(t *testing.T) {
	ev := AlertEvent{
		Source:   AlertSourceCPUTemp,
		Severity: SeverityWarning,
		Value:    91.2,
		Limit:    85,
		Message:  "CPU 91.2°C",
	}
	if got := ev.String(); got == "" {
		t.Errorf("AlertEvent.String() returned empty")
	}
}

func TestSeverityOrdering(t *testing.T) {
	if !(SeverityNormal < SeverityWarning && SeverityWarning < SeverityCritical) {
		t.Errorf("severity order broken: %d %d %d", SeverityNormal, SeverityWarning, SeverityCritical)
	}
}

type captureNotifier struct {
	got []AlertEvent
}

func (c *captureNotifier) Notify(ev AlertEvent) { c.got = append(c.got, ev) }

func TestNotifierInterfaceImplementable(t *testing.T) {
	var n Notifier = &captureNotifier{}
	n.Notify(AlertEvent{Source: AlertSourceCPUTemp})
	if cn := n.(*captureNotifier); len(cn.got) != 1 {
		t.Errorf("expected 1 event, got %d", len(cn.got))
	}
}

func TestAlerter_FiresAndRecoversWithHysteresis(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.CPUTempC = 85
	cfg.HysteresisC = 3
	cfg.CooldownMs = 0 // disable cooldown for this test
	n := &captureNotifier{}
	a := NewAlerter(cfg, n)

	a.Check(AlertSourceCPUTemp, 80) // below — no event
	a.Check(AlertSourceCPUTemp, 86) // crosses → fire
	a.Check(AlertSourceCPUTemp, 84) // still above limit-hysteresis (85-3=82) → no recover
	a.Check(AlertSourceCPUTemp, 81) // below limit-hysteresis → recover
	a.Check(AlertSourceCPUTemp, 81) // still recovered — no event

	if len(n.got) != 2 {
		t.Fatalf("expected 2 events, got %d: %v", len(n.got), n.got)
	}
	if !n.got[0].Triggered || n.got[0].Severity != SeverityWarning {
		t.Errorf("event 0 = %+v, want triggered warning", n.got[0])
	}
	if n.got[1].Triggered {
		t.Errorf("event 1 should be recover, got triggered=true")
	}
}

func TestAlerter_CooldownSuppressesRefire(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.CPUTempC = 85
	cfg.HysteresisC = 3
	cfg.CooldownMs = 30000
	n := &captureNotifier{}
	a := NewAlerter(cfg, n)

	clock := time.Unix(0, 0)
	a.now = func() time.Time { return clock }

	a.Check(AlertSourceCPUTemp, 90) // fire
	a.Check(AlertSourceCPUTemp, 80) // recover
	clock = clock.Add(5 * time.Second)
	a.Check(AlertSourceCPUTemp, 90) // within cooldown window → suppressed
	clock = clock.Add(30 * time.Second)
	a.Check(AlertSourceCPUTemp, 90) // past cooldown → fires again

	if len(n.got) != 3 {
		t.Fatalf("expected 3 events (fire, recover, fire), got %d: %+v", len(n.got), n.got)
	}
}

func TestAlerter_OnCPUMetrics_RoutesAllSources(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.CPUTempC = 50
	cfg.GPUTempC = 50
	cfg.PackagePowerW = 10
	n := &captureNotifier{}
	a := NewAlerter(cfg, n)

	a.OnCPUMetrics(CPUMetrics{CPUTemp: 90, GPUTemp: 90, PackageW: 30})

	if len(n.got) != 3 {
		t.Fatalf("expected 3 fire events, got %d: %+v", len(n.got), n.got)
	}
}

func TestAlerter_OnMemory_FiresOnPct(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.MemoryUsedPct = 80
	n := &captureNotifier{}
	a := NewAlerter(cfg, n)

	a.OnMemory(MemoryMetrics{Total: 100, Used: 90})

	if len(n.got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(n.got))
	}
	if n.got[0].Source != AlertSourceMemory {
		t.Errorf("source = %v, want memory", n.got[0].Source)
	}
}

func TestStderrNotifier_WritesEventToLogger(t *testing.T) {
	var buf strings.Builder
	logger := log.New(&buf, "", 0)
	n := &stderrNotifier{logger: logger}
	n.Notify(AlertEvent{Source: AlertSourceCPUTemp, Triggered: true, Value: 91, Limit: 85, Severity: SeverityWarning, Message: "CPU 91°C"})
	if !strings.Contains(buf.String(), "CPU 91") {
		t.Errorf("log output missing message: %q", buf.String())
	}
}

func TestAlerter_Check_PromotesToCriticalAboveCritDelta(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.CPUTempC = 85
	cfg.HysteresisC = 3
	cfg.CooldownMs = 0
	n := &captureNotifier{}
	a := NewAlerter(cfg, n)

	a.Check(AlertSourceCPUTemp, 90) // warning (above 85, below 95)
	if len(n.got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(n.got))
	}
	if n.got[0].Severity != SeverityWarning {
		t.Errorf("first event severity = %v, want Warning", n.got[0].Severity)
	}
	// Force a recover then re-cross at higher value so we re-fire on a Critical band.
	a.Check(AlertSourceCPUTemp, 80) // recover (below 85-3=82)
	a.Check(AlertSourceCPUTemp, 99) // re-fire — 99 >= 85+10 → Critical
	if len(n.got) != 3 {
		t.Fatalf("expected 3 events (warn, recover, critical), got %d: %+v", len(n.got), n.got)
	}
	if n.got[2].Severity != SeverityCritical {
		t.Errorf("third event severity = %v, want Critical", n.got[2].Severity)
	}
}

func TestAlerter_Check_PackagePowerCriticalAt125Percent(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.PackagePowerW = 40
	cfg.HysteresisPct = 5
	cfg.CooldownMs = 0
	n := &captureNotifier{}
	a := NewAlerter(cfg, n)

	a.Check(AlertSourcePackagePower, 45) // warning (above 40, below 50 = 40+25%)
	a.Check(AlertSourcePackagePower, 30) // recover (below 40-5=35)
	a.Check(AlertSourcePackagePower, 55) // re-fire — 55 >= 50 → Critical
	if len(n.got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(n.got))
	}
	if n.got[0].Severity != SeverityWarning {
		t.Errorf("first severity = %v, want Warning", n.got[0].Severity)
	}
	if n.got[2].Severity != SeverityCritical {
		t.Errorf("third severity = %v, want Critical", n.got[2].Severity)
	}
}

func TestAlerter_Check_MemoryCriticalAtLimitPlus5pp(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.MemoryUsedPct = 90
	cfg.HysteresisPct = 5
	cfg.CooldownMs = 0
	n := &captureNotifier{}
	a := NewAlerter(cfg, n)

	a.Check(AlertSourceMemory, 92) // warning (above 90, below 95)
	a.Check(AlertSourceMemory, 80) // recover (below 90-5=85)
	a.Check(AlertSourceMemory, 96) // re-fire — 96 >= 95 → Critical
	if len(n.got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(n.got))
	}
	if n.got[2].Severity != SeverityCritical {
		t.Errorf("third severity = %v, want Critical", n.got[2].Severity)
	}
}
