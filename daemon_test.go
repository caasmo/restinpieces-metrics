package metrics

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/caasmo/restinpieces/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func TestDaemon(t *testing.T) {
	var pointer atomic.Pointer[config.Config]
	pointer.Store(&config.Config{
		Metrics: config.Metrics{
			Activated:  true,
			ListenAddr: "127.0.0.1:0",
		},
	})

	recorder := NewDefaultRecorder()
	registry := prometheus.NewRegistry()
	registry.MustRegister(recorder.RequestsTotal)
	registry.MustRegister(recorder.RequestDuration)
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	var out bytes.Buffer
	daemon := NewDaemon(&pointer, slog.New(slog.NewTextHandler(&out, nil)), registry)
	if err := daemon.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	response, err := http.Get("http://" + daemon.listener.Addr().String() + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("reading the response failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET /metrics returned status %d", response.StatusCode)
	}
	if !strings.Contains(string(body), "go_goroutines") {
		t.Errorf("the metrics body does not contain go_goroutines")
	}

	line := out.String()
	if !strings.Contains(line, "http_request") {
		t.Errorf("log output does not contain http_request: %q", line)
	}
	if !strings.Contains(line, "uri=/metrics") {
		t.Errorf("log output does not contain uri=/metrics: %q", line)
	}
	if !strings.Contains(line, "status=200") {
		t.Errorf("log output does not contain status=200: %q", line)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := daemon.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}
