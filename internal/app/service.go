package app

import (
	"encoding/xml"
	"fmt"
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
