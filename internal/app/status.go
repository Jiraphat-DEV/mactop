// internal/app/status.go
package app

import (
	"fmt"
	"os"
)

// classifyStatus returns the worst severity across all monitored axes.
// Warning = above limit. Critical = above limit + 10% (temp: +10°C, power: 25%, memory: +5pp absolute).
func classifyStatus(cfg AlertsConfig, cpuTempC, gpuTempC, packageW, memPct float64) Severity {
	worst := SeverityNormal
	bump := func(s Severity) {
		if s > worst {
			worst = s
		}
	}

	check := func(v, limit, critDelta float64) Severity {
		if limit <= 0 {
			return SeverityNormal
		}
		if v >= limit+critDelta {
			return SeverityCritical
		}
		if v > limit {
			return SeverityWarning
		}
		return SeverityNormal
	}

	bump(check(cpuTempC, cfg.CPUTempC, 10))
	bump(check(gpuTempC, cfg.GPUTempC, 10))
	bump(check(packageW, cfg.PackagePowerW, cfg.PackagePowerW*0.25))
	bump(check(memPct, cfg.MemoryUsedPct, 5))

	return worst
}

func severityGlyph(s Severity) string {
	switch s {
	case SeverityCritical:
		return "🔴 Critical"
	case SeverityWarning:
		return "🟡 Warning"
	}
	return "🟢 Normal"
}

type statusSnapshot struct {
	Severity   Severity
	CPUPct     float64
	CPUTempC   float64
	GPUPct     float64
	GPUTempC   float64
	PackageW   float64
	MemoryPct  float64
	NetInKBps  float64
	NetOutKBps float64
}

func formatStatusLine(s statusSnapshot) string {
	return fmt.Sprintf(
		"%s  |  CPU %.0f%%  %.0f°C  |  GPU %.0f%%  %.0f°C  |  %.1fW  |  RAM %.0f%%  |  ↑%s ↓%s",
		severityGlyph(s.Severity),
		s.CPUPct, s.CPUTempC,
		s.GPUPct, s.GPUTempC,
		s.PackageW,
		s.MemoryPct,
		humanKBps(s.NetOutKBps),
		humanKBps(s.NetInKBps),
	)
}

func humanKBps(kbps float64) string {
	if kbps >= 1024 {
		return fmt.Sprintf("%.1fMB", kbps/1024)
	}
	return fmt.Sprintf("%.0fKB", kbps)
}

// runStatusOneLiner samples a single metric snapshot and prints a one-line
// summary, then exits. Called from runAlternateMode when --status is set.
func runStatusOneLiner() {
	if err := initSocMetrics(); err != nil {
		fmt.Fprintf(os.Stderr, "mactop status: failed to init metrics: %v\n", err)
		os.Exit(1)
	}
	defer cleanupSocMetrics()

	// Warm up CPU percent counter (delta-based; first call returns zero).
	_, _ = GetCPUPercentages()

	m := sampleSocMetrics(200)
	mem := getMemoryMetrics()
	netDisk := getNetDiskMetrics()

	cpuPercents, _ := GetCPUPercentages()
	var cpuAvg float64
	if len(cpuPercents) > 0 {
		var sum float64
		for _, p := range cpuPercents {
			sum += p
		}
		cpuAvg = sum / float64(len(cpuPercents))
	}

	var memPct float64
	if mem.Total > 0 {
		memPct = float64(mem.Used) / float64(mem.Total) * 100
	}

	cfg := ResolveAlertsConfig(currentConfig.Alerts)
	sev := classifyStatus(cfg, float64(m.CPUTemp), float64(m.GPUTemp), m.TotalPower, memPct)

	fmt.Println(formatStatusLine(statusSnapshot{
		Severity:   sev,
		CPUPct:     cpuAvg,
		CPUTempC:   float64(m.CPUTemp),
		GPUPct:     m.GPUActive,
		GPUTempC:   float64(m.GPUTemp),
		PackageW:   m.TotalPower,
		MemoryPct:  memPct,
		NetInKBps:  netDisk.InBytesPerSec / 1024,
		NetOutKBps: netDisk.OutBytesPerSec / 1024,
	}))
}
