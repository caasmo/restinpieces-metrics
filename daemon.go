package metrics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/caasmo/go-daemon-runner/daemon"
	"github.com/caasmo/restinpieces/config"
	"github.com/caasmo/restinpieces/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// readHeaderTimeout bounds reading a request header on the metrics listener.
const readHeaderTimeout = 5 * time.Second

// Daemon serves the metrics over an internal listener: loopback or a private
// address. The public application server never serves metrics.
//
// The daemon satisfies the daemon.Daemon contract: Run binds the listener and
// spawns the serving goroutine, Stop cancels the context and waits for the
// goroutine to finish. Start wraps Run for the restinpieces server.Daemon
// contract. Every scrape is logged with the same http_request message as the
// application server logs.
type Daemon struct {
	daemon.Base
	cfgPointer *atomic.Pointer[config.Config]
	registry   *prometheus.Registry

	listener net.Listener
	server   *http.Server
}

// NewDaemon creates the metrics daemon serving the given registry. The daemon
// holds the current-config box and reads metrics.listen_addr from it when it
// starts. A nil logger falls back to slog.Default().
func NewDaemon(pointer *atomic.Pointer[config.Config], logger *slog.Logger, registry *prometheus.Registry) *Daemon {
	d := &Daemon{
		Base:       daemon.NewBase("MetricsDaemon", logger),
		cfgPointer: pointer,
		registry:   registry,
	}
	d.Logger = d.Logger.With("daemon_name", d.Name())
	return d
}

// Run binds the internal listener and serves the metrics in a background
// goroutine. A bind error is a startup failure, so the server aborts startup
// instead of running without metrics. Stop cancels Ctx; the goroutine shuts
// the server down, waits for in-flight scrapes, then closes ShutdownDone.
func (d *Daemon) Run() error {
	addr := d.cfgPointer.Load().Metrics.ListenAddr

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics daemon: failed to listen on %s: %w", addr, err)
	}

	handler := promhttp.HandlerFor(d.registry, promhttp.HandlerOpts{})

	d.listener = listener
	d.server = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &core.ResponseRecorder{
				ResponseWriter: w,
				Status:         http.StatusOK,
				StartTime:      time.Now(),
			}

			handler.ServeHTTP(rec, r)

			d.Logger.Info("http_request",
				"method", r.Method,
				"uri", r.URL.RequestURI(),
				"status", rec.Status,
				"duration", rec.Duration(),
			)
		}),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	d.Logger.Info("starting", "listen_addr", listener.Addr().String())

	go func() {
		defer close(d.ShutdownDone)

		serveErr := make(chan error, 1)
		go func() {
			serveErr <- d.server.Serve(listener)
		}()

		select {
		case <-d.Ctx.Done():
			shutdownErr := d.server.Shutdown(context.Background())
			if shutdownErr != nil {
				d.Logger.Error("shutdown error", "error", shutdownErr)
			}
			<-serveErr
		case err := <-serveErr:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				d.Logger.Error("server error", "error", err)
			}
		}
	}()

	return nil
}

// Start wraps Run for the restinpieces server.Daemon lifecycle
// (Start/Stop instead of Run/Stop), so this daemon can be registered
// via srv.AddDaemon while restinpieces still manages daemons itself.
//
// TODO: remove once restinpieces is on go-daemon-runner; the runner
// calls Run directly.
func (d *Daemon) Start() error {
	return d.Run()
}
