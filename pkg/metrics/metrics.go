package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "vkino"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"service", "method", "route", "status"},
	)

	HTTPRequestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_errors_total",
			Help:      "Total number of HTTP requests with 5xx status.",
		},
		[]string{"service", "method", "route", "status"},
	)

	HTTPRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method", "route", "status"},
	)

	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests.",
		},
		[]string{"service", "method", "code"},
	)

	GRPCRequestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_request_errors_total",
			Help:      "Total number of failed gRPC requests.",
		},
		[]string{"service", "method", "code"},
	)

	GRPCRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "grpc_request_duration_seconds",
			Help:      "gRPC request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method", "code"},
	)

	GRPCStreamsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_streams_total",
			Help:      "Total number of gRPC streams.",
		},
		[]string{"service", "method", "code"},
	)

	GRPCStreamErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_stream_errors_total",
			Help:      "Total number of failed gRPC streams.",
		},
		[]string{"service", "method", "code"},
	)

	GRPCStreamDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "grpc_stream_duration_seconds",
			Help:      "gRPC stream duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method", "code"},
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
	ServiceInfo.WithLabelValues(labelValue(service, "unknown")).Set(1)
}

func labelValue(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
