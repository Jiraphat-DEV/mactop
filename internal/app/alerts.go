// internal/app/alerts.go
package app

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Severity int

const (
	SeverityNormal Severity = iota
	SeverityWarning
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityCritical:
		return "critical"
	case SeverityWarning:
		return "warning"
	default:
		return "normal"
	}
}

type AlertSource int

const (
	AlertSourceCPUTemp AlertSource = iota
	AlertSourceGPUTemp
	AlertSourcePackagePower
	AlertSourceMemory
)

func (a AlertSource) String() string {
	switch a {
	case AlertSourceCPUTemp:
		return "cpu_temp"
	case AlertSourceGPUTemp:
		return "gpu_temp"
	case AlertSourcePackagePower:
		return "package_power"
	case AlertSourceMemory:
		return "memory"
	}
	return "unknown"
}

// AlertEvent is emitted by Alerter when a metric crosses a threshold
// (Triggered=true) or recovers below threshold-hysteresis (Triggered=false).
type AlertEvent struct {
	Source    AlertSource
	Severity  Severity
	Value     float64
	Limit     float64
	Triggered bool
	Message   string
	At        time.Time
}

func (e AlertEvent) String() string {
	state := "RECOVER"
	if e.Triggered {
		state = "FIRE"
	}
	return fmt.Sprintf("[%s] %s %s=%.1f limit=%.1f", state, e.Severity, e.Source, e.Value, e.Limit)
}

// Notifier consumes AlertEvents. Implementations: stderrNotifier (Plan A),
// osascriptNotifier (Plan B), multiNotifier (fan-out).
type Notifier interface {
	Notify(ev AlertEvent)
}

// Alerter holds per-source latched state for hysteresis and cooldown.
type Alerter struct {
	mu       sync.Mutex
	cfg      AlertsConfig
	notifier Notifier
	state    map[AlertSource]alertState
	now      func() time.Time
}

type alertState struct {
	triggered bool
	lastFire  time.Time
}

func NewAlerter(cfg AlertsConfig, n Notifier) *Alerter {
	return &Alerter{
		cfg:      cfg,
		notifier: n,
		state:    make(map[AlertSource]alertState),
		now:      time.Now,
	}
}

// Check evaluates a single (source,value) pair against the configured limit.
// It emits AlertEvents through the notifier when the latch flips
// (below→above or above→below-hysteresis).
func (a *Alerter) Check(src AlertSource, value float64) {
	if !a.cfg.Enabled {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	limit, hyst := a.limitFor(src)
	if limit <= 0 {
		return
	}

	st := a.state[src]
	now := a.now()

	if !st.triggered && value > limit {
		if !st.lastFire.IsZero() && now.Sub(st.lastFire) < time.Duration(a.cfg.CooldownMs)*time.Millisecond {
			return
		}
		st.triggered = true
		st.lastFire = now
		a.state[src] = st
		a.notifier.Notify(AlertEvent{
			Source:    src,
			Severity:  SeverityWarning,
			Value:     value,
			Limit:     limit,
			Triggered: true,
			Message:   formatAlertMessage(src, value, true),
			At:        now,
		})
		return
	}

	if st.triggered && value < limit-hyst {
		st.triggered = false
		a.state[src] = st
		a.notifier.Notify(AlertEvent{
			Source:    src,
			Severity:  SeverityNormal,
			Value:     value,
			Limit:     limit,
			Triggered: false,
			Message:   formatAlertMessage(src, value, false),
			At:        now,
		})
	}
}

func (a *Alerter) limitFor(src AlertSource) (limit, hyst float64) {
	switch src {
	case AlertSourceCPUTemp:
		return a.cfg.CPUTempC, a.cfg.HysteresisC
	case AlertSourceGPUTemp:
		return a.cfg.GPUTempC, a.cfg.HysteresisC
	case AlertSourcePackagePower:
		return a.cfg.PackagePowerW, a.cfg.HysteresisPct // reused as watt-window
	case AlertSourceMemory:
		return a.cfg.MemoryUsedPct, a.cfg.HysteresisPct
	}
	return 0, 0
}

func formatAlertMessage(src AlertSource, value float64, fired bool) string {
	verb := "recovered"
	if fired {
		verb = "exceeded"
	}
	switch src {
	case AlertSourceCPUTemp:
		return fmt.Sprintf("CPU temperature %s: %.1f°C", verb, value)
	case AlertSourceGPUTemp:
		return fmt.Sprintf("GPU temperature %s: %.1f°C", verb, value)
	case AlertSourcePackagePower:
		return fmt.Sprintf("Package power %s: %.1fW", verb, value)
	case AlertSourceMemory:
		return fmt.Sprintf("Memory usage %s: %.1f%%", verb, value)
	}
	return ""
}

// OnCPUMetrics inspects the CPU/GPU temp and package power fields of a
// CPUMetrics snapshot (the same struct used by the existing metrics loop)
// and routes each to Check.
func (a *Alerter) OnCPUMetrics(m CPUMetrics) {
	a.Check(AlertSourceCPUTemp, m.CPUTemp)
	a.Check(AlertSourceGPUTemp, m.GPUTemp)
	a.Check(AlertSourcePackagePower, m.PackageW)
}

// OnMemory computes used-percent from MemoryMetrics and routes to Check.
func (a *Alerter) OnMemory(m MemoryMetrics) {
	if m.Total == 0 {
		return
	}
	pct := float64(m.Used) / float64(m.Total) * 100.0
	a.Check(AlertSourceMemory, pct)
}

type stderrNotifier struct {
	logger *log.Logger
}

func newStderrNotifier(logger *log.Logger) *stderrNotifier {
	return &stderrNotifier{logger: logger}
}

func (s *stderrNotifier) Notify(ev AlertEvent) {
	s.logger.Printf("alert: %s", ev.Message)
}
