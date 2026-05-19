package app

import (
	"fmt"
	"strings"
)

// osascriptNotifier delivers AlertEvents as native macOS banners via
// /usr/bin/osascript. By default it stays silent on recovery events
// to avoid double-noise (banner-on, banner-off); set noisyRecovery to
// emit recovery banners too.
type osascriptNotifier struct {
	runner        cmdRunner
	appName       string
	noisyRecovery bool
}

func newOsascriptNotifier() *osascriptNotifier {
	return &osascriptNotifier{runner: execRunner{}, appName: "mactop"}
}

func (o *osascriptNotifier) Notify(ev AlertEvent) {
	if !ev.Triggered && !o.noisyRecovery {
		return
	}
	title := "mactop"
	if !ev.Triggered {
		title = "mactop — recovered"
	}
	escTitle := escapeAppleScript(title)
	escMsg := escapeAppleScript(ev.Message)
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, escMsg, escTitle)
	_ = o.runner.Run("/usr/bin/osascript", "-e", script)
}

func escapeAppleScript(s string) string {
	// AppleScript string literals: escape backslash, then double quote.
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
