package metrics

import (
	"io"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/caasmo/restinpieces/config"
)

func TestNewDefault(t *testing.T) {
	var pointer atomic.Pointer[config.Config]
	pointer.Store(&config.Config{
		Metrics: config.Metrics{
			Activated:  true,
			ListenAddr: "127.0.0.1:0",
		},
	})

	recorder, daemon := NewDefault(&pointer, slog.New(slog.NewTextHandler(io.Discard, nil)))

	if recorder == nil {
		t.Fatal("NewDefault() returned a nil recorder")
	}
	if daemon == nil {
		t.Fatal("NewDefault() returned a nil daemon")
	}
}
