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
	registry.MustRegister(SegmentCenterFrequency)
	registry.MustRegister(SegmentSampleRate)
	registry.MustRegister(WebConnections)
	registry.MustRegister(MaxConnections)
	registry.MustRegister(MaxWebConnections)
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
	MaxConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "max_connections",
		Help: "The max concurrent connections this server accepts",
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
		Subsystem: "server",
		Name:      "center_frequency",
		Help:      "Server Center Frequency in Hertz",
	})
	ServerSampleRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "server",
		Name:      "samplerate",
		Help:      "Server Sample Rate in Samples Per Second",
	})
	SegmentCenterFrequency = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "segment",
		Name:      "center_frequency",
		Help:      "Beacon Segment Center Frequency in Hertz",
	})
	SegmentSampleRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "segment",
		Name:      "samplerate",
		Help:      "Beacon Segment Rate in Samples Per Second",
	})
	WebConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "server",
		Name:      "webconnections",
		Help:      "Current WebSocket Connections",
	})
	MaxWebConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "max_web_connections",
		Help: "The max concurrent connections to websocket this server accepts",
	})
)

func GetHandler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
