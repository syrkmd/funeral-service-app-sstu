package repository

import (
	"context"
	"strconv"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusRepository struct {
	totalRequests        prometheus.Counter
	requestsByStatus     *prometheus.CounterVec
	requestDuration      *prometheus.HistogramVec
	activeConnections    prometheus.Gauge
	blockedRequests      prometheus.Counter
	rateLimitedRequests  prometheus.Counter
	connectionViolations *prometheus.CounterVec
	bandwidthViolations  *prometheus.CounterVec
	upstreamErrors       prometheus.Counter
	upstreamLatency      *prometheus.HistogramVec
	bytesIn              prometheus.Counter
	bytesOut             prometheus.Counter
	requestsByMethod     *prometheus.CounterVec
	requestsByPath       *prometheus.CounterVec
	cacheEvents          *prometheus.CounterVec
}

func NewPrometheusRepository(registerer prometheus.Registerer) *PrometheusRepository {
	factory := promauto.With(registerer)

	return &PrometheusRepository{
		totalRequests: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "requests_total",
			Help:      "Total number of completed requests.",
		}),
		requestsByStatus: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "requests_by_status_total",
			Help:      "Total number of requests grouped by response status.",
		}, []string{"status"}),
		requestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "proxy",
			Name:      "request_duration_seconds",
			Help:      "Request latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path", "status"}),
		activeConnections: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "proxy",
			Name:      "active_connections",
			Help:      "Current number of in-flight requests.",
		}),
		blockedRequests: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "blocked_requests_total",
			Help:      "Total number of blocked requests.",
		}),
		rateLimitedRequests: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "rate_limited_requests_total",
			Help:      "Total number of rate-limited requests.",
		}),
		connectionViolations: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "connection_limit_violations_total",
			Help:      "Total number of connection-related rate limit violations.",
		}, []string{"type"}),
		bandwidthViolations: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "bandwidth_limit_violations_total",
			Help:      "Total number of bandwidth-related rate limit violations.",
		}, []string{"type"}),
		upstreamErrors: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "upstream_errors_total",
			Help:      "Total number of upstream errors.",
		}),
		upstreamLatency: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "proxy",
			Name:      "upstream_latency_seconds",
			Help:      "Observed upstream latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"path", "status"}),
		bytesIn: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "bytes_in_total",
			Help:      "Total inbound bytes received by the proxy.",
		}),
		bytesOut: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "bytes_out_total",
			Help:      "Total outbound bytes sent by the proxy.",
		}),
		requestsByMethod: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "requests_by_method_total",
			Help:      "Total number of requests grouped by HTTP method.",
		}, []string{"method"}),
		requestsByPath: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "requests_by_path_total",
			Help:      "Total number of requests grouped by normalized path.",
		}, []string{"path"}),
		cacheEvents: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "proxy",
			Name:      "cache_events_total",
			Help:      "Total number of cache events grouped by action.",
		}, []string{"action"}),
	}
}

func (r *PrometheusRepository) BeginRequest(_ context.Context) {
	r.activeConnections.Inc()
}

func (r *PrometheusRepository) FinishRequest(_ context.Context, record domain.RequestRecord) {
	status := strconv.Itoa(record.Status)
	r.totalRequests.Inc()
	r.requestsByStatus.WithLabelValues(status).Inc()
	r.requestsByMethod.WithLabelValues(record.Method).Inc()
	r.requestsByPath.WithLabelValues(record.Path).Inc()
	r.requestDuration.WithLabelValues(record.Method, record.Path, status).Observe(record.Latency.Seconds())
	r.activeConnections.Dec()

	if record.BytesIn > 0 {
		r.bytesIn.Add(float64(record.BytesIn))
	}
	if record.BytesOut > 0 {
		r.bytesOut.Add(float64(record.BytesOut))
	}
}

func (r *PrometheusRepository) RecordAccessDecision(_ context.Context, event domain.AccessDecisionEvent) {
	if event.Decision == "deny" {
		r.blockedRequests.Inc()
	}
}

func (r *PrometheusRepository) RecordRateLimitDecision(_ context.Context, event domain.RateLimitDecisionEvent) {
	if !event.Allowed {
		r.blockedRequests.Inc()
		r.rateLimitedRequests.Inc()
		if event.LimitType == "connections" || event.LimitType == "cps" {
			r.connectionViolations.WithLabelValues(event.LimitType).Inc()
		}
		if event.LimitType == "upload" || event.LimitType == "download" || event.LimitType == "total" {
			r.bandwidthViolations.WithLabelValues(event.LimitType).Inc()
		}
	}
}

func (r *PrometheusRepository) RecordUpstream(_ context.Context, event domain.UpstreamEvent) {
	status := strconv.Itoa(event.StatusCode)
	r.upstreamLatency.WithLabelValues(event.Path, status).Observe(event.Latency.Seconds())
	if event.Error != "" {
		r.upstreamErrors.Inc()
	}
}

func (r *PrometheusRepository) RecordCacheEvent(_ context.Context, event domain.CacheEvent) {
	r.cacheEvents.WithLabelValues(event.Action).Inc()
}
