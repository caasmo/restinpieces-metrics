# restinpieces-metrics

The default Prometheus metrics for a restinpieces application, and the internal daemon that serves them. The framework's metrics middleware hands every finished response to the `core.MetricsRecorder` set on the app; this repository provides the collectors and the daemon.

## Use the default metrics

```go
recorder, daemon := metrics.NewDefault(coreApp.ConfigPointer(), coreApp.Logger())
coreApp.SetMetrics(recorder)
srv.AddDaemon(daemon)
```

The default recorder counts requests (`http_server_requests_total`) and records request time (`http_server_request_duration_seconds`), both labeled by status code. The registry also carries the Go and process collectors, and the daemon serves them over HTTP on `metrics.listen_addr`, which is loopback or private.

## Use your own collectors

Write a type with a `Record(*core.ResponseRecorder)` method, register its collectors together with the Go and process collectors in a private registry, and pass the registry to `NewDaemon`:

```go
type myMetrics struct {
	RequestsTotal *prometheus.CounterVec
}

func (m *myMetrics) Record(rec *core.ResponseRecorder) {
	m.RequestsTotal.WithLabelValues(strconv.Itoa(rec.Status)).Inc()
}

m := &myMetrics{
	RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "my_requests_total",
		Help: "Requests handled by my application.",
	}, []string{"code"}),
}

registry := prometheus.NewRegistry()
registry.MustRegister(m.RequestsTotal)
registry.MustRegister(collectors.NewGoCollector())
registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

coreApp.SetMetrics(m)
srv.AddDaemon(metrics.NewDaemon(coreApp.ConfigPointer(), logger, registry))
```

Set the recorder on the app before the server starts.
