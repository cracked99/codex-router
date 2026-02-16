package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
)

var (
	requestCount    atomic.Int64
	errorCount      atomic.Int64
	totalLatencyMs atomic.Int64
)

// MetricsHandler returns Prometheus-style metrics
func MetricsHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		reqs := requestCount.Load()
		errs := errorCount.Load()
		latency := totalLatencyMs.Load()

		var avgLatency float64
		if reqs > 0 {
			avgLatency = float64(latency) / float64(reqs)
		}

		metrics := `# HELP codex_router_requests_total Total number of requests
# TYPE codex_router_requests_total counter
codex_router_requests_total ` + fmt.Sprint(reqs) + `

# HELP codex_router_errors_total Total number of errors
# TYPE codex_router_errors_total counter
codex_router_errors_total ` + fmt.Sprint(errs) + `

# HELP codex_router_latency_avg_ms Average request latency in milliseconds
# TYPE codex_router_latency_avg_ms gauge
codex_router_latency_avg_ms ` + fmt.Sprintf("%.2f", avgLatency) + `

# HELP codex_router_up Server is up
# TYPE codex_router_up gauge
codex_router_up 1
`

		w.Write([]byte(metrics))
	}
}
