package app

import (
	"strings"
	"testing"
)

func TestOsascriptNotifier_BuildsValidScript(t *testing.T) {
	r := &fakeRunner{}
	n := &osascriptNotifier{runner: r, appName: "mactop"}
	n.Notify(AlertEvent{Source: AlertSourceCPUTemp, Triggered: true, Value: 91.2, Limit: 85, Message: "CPU 91°C"})

	if len(r.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(r.calls))
	}
	if r.calls[0][0] != "/usr/bin/osascript" {
		t.Errorf("command = %q, want /usr/bin/osascript", r.calls[0][0])
	}
	joined := strings.Join(r.calls[0], " ")
	if !strings.Contains(joined, "display notification") {
		t.Errorf("script missing 'display notification': %s", joined)
	}
	if !strings.Contains(joined, "CPU 91°C") {
		t.Errorf("script missing alert message")
	}
}

func TestOsascriptNotifier_EscapesQuotes(t *testing.T) {
	r := &fakeRunner{}
	n := &osascriptNotifier{runner: r, appName: "mactop"}
	n.Notify(AlertEvent{Triggered: true, Message: `He said "hot"`})

	joined := strings.Join(r.calls[0], " ")
	// AppleScript backslash-escapes double quotes inside string literals.
	if !strings.Contains(joined, `\"hot\"`) {
		t.Errorf("quotes not escaped:\n%s", joined)
	}
}

func TestOsascriptNotifier_SkipsRecoveryByDefault(t *testing.T) {
	r := &fakeRunner{}
	n := &osascriptNotifier{runner: r, appName: "mactop"}
	n.Notify(AlertEvent{Triggered: false, Message: "ok"})
	if len(r.calls) != 0 {
		t.Errorf("expected 0 calls for non-noisy recovery, got %d", len(r.calls))
	}
}

func TestOsascriptNotifier_RecoveryEmittedWhenNoisyRecoveryEnabled(t *testing.T) {
	r := &fakeRunner{}
	n := &osascriptNotifier{runner: r, appName: "mactop", noisyRecovery: true}
	n.Notify(AlertEvent{Triggered: false, Message: "ok"})
	if len(r.calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(r.calls))
	}
}
