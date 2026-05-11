package app

import (
	"strings"
	"testing"

	"github.com/metaspartan/mactop/v2/internal/i18n"
)

func init() { i18n.Init("en") }

func TestClassifyStatus(t *testing.T) {
	cfg := DefaultAlertsConfig()
	cfg.CPUTempC = 85
	cfg.GPUTempC = 85
	cfg.PackagePowerW = 40
	cfg.MemoryUsedPct = 90

	tests := []struct {
		name string
		cpu  float64
		gpu  float64
		pkg  float64
		mem  float64
		want Severity
	}{
		{"all normal", 50, 40, 10, 50, SeverityNormal},
		{"cpu temp warns", 86, 40, 10, 50, SeverityWarning},
		{"power warns", 50, 40, 41, 50, SeverityWarning},
		{"memory critical at +10pct", 50, 40, 10, 99, SeverityCritical},
		{"cpu temp critical at +10°C", 96, 40, 10, 50, SeverityCritical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyStatus(cfg, tt.cpu, tt.gpu, tt.pkg, tt.mem)
			if got != tt.want {
				t.Errorf("classifyStatus = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatStatusLine(t *testing.T) {
	snap := statusSnapshot{
		Severity:   SeverityNormal,
		CPUPct:     14,
		CPUTempC:   45,
		GPUPct:     8,
		GPUTempC:   41,
		PackageW:   8.2,
		MemoryPct:  61,
		NetInKBps:  340,
		NetOutKBps: 1228, // → 1.2MB
	}
	got := formatStatusLine(snap)

	for _, want := range []string{
		"🟢 Normal",
		"CPU 14%", "45°C",
		"GPU 8%", "41°C",
		"8.2W",
		"RAM 61%",
		"↑1.2MB", "↓340KB",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("formatStatusLine missing %q in %q", want, got)
		}
	}
}
