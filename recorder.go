// Package metrics provides the default Prometheus collectors for a
// restinpieces application and the internal daemon that serves them.
//
// The application has two options. NewDefault builds the default collectors,
// the registry and the daemon in one call. An application that wants
// different collectors builds its own recorder with a Record method,
// registers them in a registry, and passes the registry to NewDaemon.
package metrics

import (
	"strconv"

	"github.com/caasmo/restinpieces/core"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	requestsTotal   = "http_server_requests_total"
	requestDuration = "http_server_request_duration_seconds"
)

// requestDurationBuckets covers the framework's fast handlers, from 10
// microseconds to 500 milliseconds; slower responses land in the +Inf
// bucket. Prometheus requires the base unit to be seconds, so the boundaries
// are in seconds: 0.00001 is 10µs, 0.001 is 1ms and 0.1 is 100ms.
var requestDurationBuckets = []float64{
	0.00001,
	0.00005,
	0.00025,
	0.001,
	0.005,
	0.025,
	0.1,
	0.5,
}

// DefaultRecorder holds the default HTTP request collectors and implements
// core.MetricsRecorder, so an application can set it on the app with
// SetMetrics and read the same collectors for its own handler
// metrics.
type DefaultRecorder struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

// NewDefaultRecorder builds the default request counter and request-time
// histogram.
func NewDefaultRecorder() *DefaultRecorder {
	return &DefaultRecorder{
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: requestsTotal,
			Help: "Total number of HTTP requests handled by the server, labeled by status code.",
		}, []string{"code"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    requestDuration,
			Help:    "HTTP request time in seconds, labeled by status code.",
			Buckets: requestDurationBuckets,
		}, []string{"code"}),
	}
}

// Record records one finished response: it counts the request and adds the
// request time to the histogram, both labeled by status code.
func (r *DefaultRecorder) Record(rec *core.ResponseRecorder) {
	status := strconv.Itoa(rec.Status)
	r.RequestsTotal.WithLabelValues(status).Inc()
	r.RequestDuration.WithLabelValues(status).Observe(rec.Duration().Seconds())
}
