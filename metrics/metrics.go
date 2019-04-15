package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var registry = prometheus.NewRegistry()

func init() {
	registry.MustRegister(Connections)
	registry.MustRegister(TotalConnections)
	registry.MustRegister(BytesOut)
	registry.MustRegister(BytesIn)
	registry.MustRegister(LockOffset)
	registry.MustRegister(ServerCenterFrequency)
	registry.MustRegister(ServerSampleRate)
}

var (
	Connections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "connections",
		Help: "Current number of connections",
	})
	TotalConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_connections",
		Help: "The total number of connections since server started",
	})
	BytesOut = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bytes_out",
		Help: "Number of bytes sent",
	})
	BytesIn = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bytes_in",
		Help: "Number of bytes received",
	})
	LockOffset = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "lock_offset",
		Help: "Offset Frequency in Hertz of the current beacon lock",
	})
	ServerCenterFrequency = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "server_center_frequency",
		Help: "Server Center Frequency in Hertz",
	})
	ServerSampleRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "server_samplerate",
		Help: "Server Sample Rate in Samples Per Second",
	})
)

func GetHandler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
