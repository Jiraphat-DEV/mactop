// internal/app/status.go
package app

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
