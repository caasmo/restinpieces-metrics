package metrics

import (
	"log/slog"
	"sync/atomic"

	"github.com/caasmo/restinpieces/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// NewDefault builds the default request metrics, registers them together
// with the Go and process collectors in a private registry, and returns the
// recorder and the daemon. Set the recorder on the app with core.App.SetMetrics
// and add the daemon to the server with Server.AddDaemon.
//
// Example:
//
//	recorder, daemon := metrics.NewDefault(coreApp.ConfigPointer(), coreApp.Logger())
//	coreApp.SetMetrics(recorder)
//	srv.AddDaemon(daemon)
func NewDefault(pointer *atomic.Pointer[config.Config], logger *slog.Logger) (*DefaultRecorder, *Daemon) {
	recorder := NewDefaultRecorder()

	registry := prometheus.NewRegistry()
	registry.MustRegister(recorder.RequestsTotal)
	registry.MustRegister(recorder.RequestDuration)
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	daemon := NewDaemon(pointer, logger, registry)

	return recorder, daemon
}
