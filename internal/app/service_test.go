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

func TestInstallPaths_ReturnsHomeBasedPaths(t *testing.T) {
	got, err := installPathsForHome("/Users/alice")
	if err != nil {
		t.Fatal(err)
	}
	if got.PlistPath != "/Users/alice/Library/LaunchAgents/com.metaspartan.mactop.plist" {
		t.Errorf("PlistPath = %q", got.PlistPath)
	}
	if got.LogPath != "/Users/alice/.mactop/daemon.log" {
		t.Errorf("LogPath = %q", got.LogPath)
	}
}

func TestInstallPaths_RejectsEmptyHome(t *testing.T) {
	if _, err := installPathsForHome(""); err == nil {
		t.Error("expected error for empty home")
	}
}
