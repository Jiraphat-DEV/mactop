package app

import (
	"strings"
	"testing"
)

func TestRenderPlist_ContainsRequiredKeys(t *testing.T) {
	got := renderPlist("/usr/local/bin/mactop", "/Users/alice/.mactop/daemon.log")
	for _, want := range []string{
		"<key>Label</key>",
		"<string>com.metaspartan.mactop</string>",
		"<key>ProgramArguments</key>",
		"<string>/usr/local/bin/mactop</string>",
		"<string>--daemon</string>",
		"<key>RunAtLoad</key>",
		"<true/>",
		"<key>KeepAlive</key>",
		"<key>StandardOutPath</key>",
		"<string>/Users/alice/.mactop/daemon.log</string>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("renderPlist missing %q\n--- output ---\n%s", want, got)
		}
	}
}

func TestRenderPlist_EscapesAmpersands(t *testing.T) {
	got := renderPlist("/Applications/My&App/mactop", "/tmp/log")
	if !strings.Contains(got, "/Applications/My&amp;App/mactop") {
		t.Errorf("renderPlist did not XML-escape '&' in path")
	}
}
