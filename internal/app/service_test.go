package app

import (
	"os"
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
		"<key>ThrottleInterval</key>",
		"<integer>15</integer>",
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

type fakeRunner struct {
	calls [][]string
	err   error
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, append([]string{name}, args...))
	return f.err
}

func TestLaunchctlBootstrap_CallsLoad(t *testing.T) {
	r := &fakeRunner{}
	if err := launchctlBootstrap(r, "/path/to.plist", 501); err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(r.calls))
	}
	want := []string{"/bin/launchctl", "bootstrap", "gui/501", "/path/to.plist"}
	for i, w := range want {
		if r.calls[0][i] != w {
			t.Errorf("call arg[%d] = %q, want %q", i, r.calls[0][i], w)
		}
	}
}

func TestLaunchctlBootout_CallsBootout(t *testing.T) {
	r := &fakeRunner{}
	if err := launchctlBootout(r, "/path/to.plist", 501); err != nil {
		t.Fatal(err)
	}
	want := []string{"/bin/launchctl", "bootout", "gui/501", "/path/to.plist"}
	for i, w := range want {
		if r.calls[0][i] != w {
			t.Errorf("call arg[%d] = %q, want %q", i, r.calls[0][i], w)
		}
	}
}

func TestInstall_WritesPlistAndCallsBootstrap(t *testing.T) {
	tmp := t.TempDir()
	r := &fakeRunner{}

	err := installService(installOptions{
		ExecPath: "/usr/local/bin/mactop",
		Home:     tmp,
		UID:      501,
		Runner:   r,
	})
	if err != nil {
		t.Fatal(err)
	}

	plistPath := tmp + "/Library/LaunchAgents/com.metaspartan.mactop.plist"
	if _, err := os.Stat(plistPath); err != nil {
		t.Fatalf("plist not written: %v", err)
	}

	if len(r.calls) != 2 || r.calls[1][1] != "bootstrap" {
		t.Errorf("expected bootout then bootstrap calls, got %v", r.calls)
	}
}

func TestInstall_OverwritesExistingPlist(t *testing.T) {
	tmp := t.TempDir()
	r := &fakeRunner{}
	opt := installOptions{ExecPath: "/old/path/mactop", Home: tmp, UID: 501, Runner: r}
	if err := installService(opt); err != nil {
		t.Fatal(err)
	}
	opt.ExecPath = "/new/path/mactop"
	if err := installService(opt); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(tmp + "/Library/LaunchAgents/com.metaspartan.mactop.plist")
	if !strings.Contains(string(data), "/new/path/mactop") {
		t.Errorf("plist not overwritten with new exec path")
	}
}

func TestUninstall_RemovesPlistAndCallsBootout(t *testing.T) {
	tmp := t.TempDir()
	r := &fakeRunner{}
	opt := installOptions{ExecPath: "/x/mactop", Home: tmp, UID: 501, Runner: r}
	if err := installService(opt); err != nil {
		t.Fatal(err)
	}
	r.calls = nil

	if err := uninstallService(uninstallOptions{Home: tmp, UID: 501, Runner: r}); err != nil {
		t.Fatal(err)
	}

	plistPath := tmp + "/Library/LaunchAgents/com.metaspartan.mactop.plist"
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Errorf("plist still exists")
	}
	if len(r.calls) != 1 || r.calls[0][1] != "bootout" {
		t.Errorf("expected bootout call, got %v", r.calls)
	}
}

func TestUninstall_NoErrorIfNotInstalled(t *testing.T) {
	r := &fakeRunner{}
	tmp := t.TempDir()
	if err := uninstallService(uninstallOptions{Home: tmp, UID: 501, Runner: r}); err != nil {
		t.Errorf("expected nil error on missing install, got %v", err)
	}
}
