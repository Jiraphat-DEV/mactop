package app

import "testing"

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
