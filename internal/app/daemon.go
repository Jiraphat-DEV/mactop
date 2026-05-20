// internal/app/daemon.go
package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/metaspartan/mactop/v2/internal/i18n"
)

// runDaemon is the entry point for `mactop --daemon`. It runs metric
// collection in the background, feeds the Alerter, and never touches a
// terminal. Designed to be invoked by launchd.
func runDaemon() {
	if err := initSocMetrics(); err != nil {
		stderrLogger.Fatalf(i18n.T("Headless_ErrorInitMetrics"), err)
	}
	defer cleanupSocMetrics()

	cfg := ResolveAlertsConfig(currentConfig.Alerts)
	alerter = NewAlerter(cfg, newOsascriptNotifier())

	// Warm the CPU percent delta source.
	_, _ = GetCPUPercentages()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(updateInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			return
		case <-ticker.C:
			m := sampleSocMetrics(updateInterval)
			mem := getMemoryMetrics()

			componentSum := m.TotalPower
			totalPower := m.SystemPower
			if totalPower < componentSum {
				totalPower = componentSum
			}

			alerter.OnCPUMetrics(CPUMetrics{
				CPUTemp:  float64(m.CPUTemp),
				GPUTemp:  float64(m.GPUTemp),
				PackageW: totalPower,
			})
			alerter.OnMemory(mem)
		}
	}
}
