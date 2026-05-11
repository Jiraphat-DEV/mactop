// internal/app/alerts.go
package app

import (
	"fmt"
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
