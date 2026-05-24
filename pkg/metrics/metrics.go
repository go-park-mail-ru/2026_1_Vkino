package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelService = "service"
	labelMethod  = "method"
	labelRoute   = "route"
	labelStatus  = "status"
	labelCode    = "code"

	namespace         = "vkino"
	unknownLabelValue = "unknown"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{labelService, labelMethod, labelRoute, labelStatus},
	)

	HTTPRequestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_errors_total",
			Help:      "Total number of HTTP requests with 5xx status.",
		},
		[]string{labelService, labelMethod, labelRoute, labelStatus},
	)

	HTTPRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{labelService, labelMethod, labelRoute, labelStatus},
	)

	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests.",
		},
		[]string{labelService, labelMethod, labelCode},
	)

	GRPCRequestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_request_errors_total",
			Help:      "Total number of failed gRPC requests.",
		},
		[]string{labelService, labelMethod, labelCode},
	)

	GRPCRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "grpc_request_duration_seconds",
			Help:      "gRPC request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{labelService, labelMethod, labelCode},
	)

	GRPCStreamsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_streams_total",
			Help:      "Total number of gRPC streams.",
		},
		[]string{labelService, labelMethod, labelCode},
	)

	GRPCStreamErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_stream_errors_total",
			Help:      "Total number of failed gRPC streams.",
		},
		[]string{labelService, labelMethod, labelCode},
	)

	GRPCStreamDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "grpc_stream_duration_seconds",
			Help:      "gRPC stream duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{labelService, labelMethod, labelCode},
	)

	ServiceInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_info",
			Help:      "Static information about the running service.",
		},
		[]string{"service"},
	)

	registerOnce sync.Once
)

func Register() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			HTTPRequestsTotal,
			HTTPRequestErrorsTotal,
			HTTPRequestDurationSeconds,
			GRPCRequestsTotal,
			GRPCRequestErrorsTotal,
			GRPCRequestDurationSeconds,
			GRPCStreamsTotal,
			GRPCStreamErrorsTotal,
			GRPCStreamDurationSeconds,
			ServiceInfo,
		)
	})
}

func SetServiceInfo(service string) {
	Register()
	ServiceInfo.WithLabelValues(labelValue(service)).Set(1)
}

func labelValue(value string) string {
	if value == "" {
		return unknownLabelValue
	}

	return value
}
