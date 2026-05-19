package app

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const serviceLabel = "com.metaspartan.mactop"

func renderPlist(execPath, logPath string) string {
	// We hand-format the plist rather than encoding/plist (not stdlib) to keep
	// zero deps. xml.EscapeText handles entity escaping for paths with '&'.
	esc := func(s string) string {
		var b strings.Builder
		_ = xml.EscapeText(&b, []byte(s))
		return b.String()
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>--daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
    <key>ProcessType</key>
    <string>Background</string>
</dict>
</plist>
`, serviceLabel, esc(execPath), esc(logPath), esc(logPath))
}

type installPaths struct {
	PlistPath string
	LogPath   string
	HomeDir   string
}

func installPathsForHome(home string) (installPaths, error) {
	if home == "" {
		return installPaths{}, errors.New("home directory not resolved")
	}
	return installPaths{
		PlistPath: filepath.Join(home, "Library", "LaunchAgents", serviceLabel+".plist"),
		LogPath:   filepath.Join(home, ".mactop", "daemon.log"),
		HomeDir:   home,
	}, nil
}

// cmdRunner abstracts os/exec so tests don't shell out.
type cmdRunner interface {
	Run(name string, args ...string) error
}

type execRunner struct{}

func (execRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

func launchctlBootstrap(r cmdRunner, plistPath string, uid int) error {
	return r.Run("/bin/launchctl", "bootstrap", fmt.Sprintf("gui/%d", uid), plistPath)
}

func launchctlBootout(r cmdRunner, plistPath string, uid int) error {
	return r.Run("/bin/launchctl", "bootout", fmt.Sprintf("gui/%d", uid), plistPath)
}

type installOptions struct {
	ExecPath string
	Home     string
	UID      int
	Runner   cmdRunner
}

func installService(opt installOptions) error {
	paths, err := installPathsForHome(opt.Home)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.PlistPath), 0755); err != nil {
		return fmt.Errorf("mkdir LaunchAgents: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(paths.LogPath), 0755); err != nil {
		return fmt.Errorf("mkdir .mactop: %w", err)
	}

	contents := renderPlist(opt.ExecPath, paths.LogPath)
	if err := os.WriteFile(paths.PlistPath, []byte(contents), 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	// Best-effort: ignore "already loaded" errors by attempting bootout first.
	_ = launchctlBootout(opt.Runner, paths.PlistPath, opt.UID)
	return launchctlBootstrap(opt.Runner, paths.PlistPath, opt.UID)
}

type uninstallOptions struct {
	Home   string
	UID    int
	Runner cmdRunner
}

func uninstallService(opt uninstallOptions) error {
	paths, err := installPathsForHome(opt.Home)
	if err != nil {
		return err
	}
	// bootout first (ignore failure; service may already be down)
	_ = launchctlBootout(opt.Runner, paths.PlistPath, opt.UID)

	if err := os.Remove(paths.PlistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist: %w", err)
	}
	return nil
}
