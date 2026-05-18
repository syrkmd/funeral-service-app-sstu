package gintransport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

type ControllerProxyUseCase interface {
	ResolveUpstream(ctx context.Context) (*url.URL, error)
	FlushInterval() int64
}

type ControllerCacheUseCase interface {
	Lookup(ctx context.Context, request domain.CacheRequest) (domain.CacheEntry, bool, error)
	Set(ctx context.Context, candidate domain.CacheCandidate, body []byte) (domain.CacheEntry, bool, error)
	DecideCachePolicy(input domain.CacheCandidate) domain.CacheDecision
	InvalidateByKey(ctx context.Context, key string) (int, error)
	InvalidateByPrefix(ctx context.Context, prefix string) (int, error)
	InvalidateByRegex(ctx context.Context, pattern string) (int, error)
	InvalidateByTags(ctx context.Context, tags []string) (int, error)
	InvalidateExpired(ctx context.Context) (int, error)
	Clear(ctx context.Context) (int, error)
	HandleInvalidationEvent(ctx context.Context, event domain.CacheInvalidationEvent) (int, error)
}

type ProxyController struct {
	useCase    ControllerProxyUseCase
	monitoring usecase.MonitoringRecorder
	rateLimit  rateLimitBandwidthController
	cache      cacheProxyController
}

const (
	proxyContextRateLimitDecisionKey     = "rate_limit_decision"
	proxyContextReservedDownloadBytesKey = "rate_limit_reserved_download_bytes"
	proxyContextCacheStatusKey           = "cache_status"
)

type rateLimitBandwidthController interface {
	ReserveResponseBandwidth(ctx context.Context, rawIP string, downloadBytes int64) (domain.RateLimitDecision, error)
}

type cacheProxyController interface {
	Lookup(ctx context.Context, request domain.CacheRequest) (domain.CacheEntry, bool, error)
	Set(ctx context.Context, candidate domain.CacheCandidate, body []byte) (domain.CacheEntry, bool, error)
	DecideCachePolicy(input domain.CacheCandidate) domain.CacheDecision
}

type responseBandwidthLimitError struct {
	decision domain.RateLimitDecision
}

func (e *responseBandwidthLimitError) Error() string {
	return "rate limit exceeded"
}

func NewProxyController(useCase ControllerProxyUseCase, monitoring usecase.MonitoringRecorder, rateLimit rateLimitBandwidthController, cache cacheProxyController) *ProxyController {
	return &ProxyController{useCase: useCase, monitoring: monitoring, rateLimit: rateLimit, cache: cache}
}

func (h *ProxyController) Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheRequest := buildCacheRequest(c.Request)
		if cacheEntry, ok := h.serveCachedResponse(c, cacheRequest); ok {
			_ = cacheEntry
			return
		}

		upstream, err := h.useCase.ResolveUpstream(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		startedAt := time.Now()

		proxy := &httputil.ReverseProxy{
			ModifyResponse: func(response *http.Response) error {
				upstreamLatency := time.Since(startedAt)
				c.Set("upstream_latency", upstreamLatency)
				h.monitoring.RecordUpstream(c.Request.Context(), usecase.RecordUpstreamInput{
					Path:       c.Request.URL.Path,
					StatusCode: response.StatusCode,
					Latency:    upstreamLatency,
				})
				if response.ContentLength > 0 {
					decision, err := h.rateLimit.ReserveResponseBandwidth(c.Request.Context(), c.ClientIP(), response.ContentLength)
					if err != nil {
						return err
					}
					if !decision.Allowed {
						c.Set(proxyContextRateLimitDecisionKey, "deny")
						h.monitoring.RecordRateLimitDecision(c.Request.Context(), usecase.RecordRateLimitInput{
							IP:           decision.IP,
							RuleID:       decision.RuleID,
							LimitType:    decision.LimitType,
							Allowed:      decision.Allowed,
							CurrentValue: decision.CurrentValue,
						})
						return &responseBandwidthLimitError{decision: decision}
					}
					c.Set(proxyContextReservedDownloadBytesKey, response.ContentLength)
				}

				if err := h.captureAndStoreResponse(c, cacheRequest, response); err != nil {
					return err
				}
				return nil
			},
			Rewrite: func(req *httputil.ProxyRequest) {
				req.SetURL(upstream)
				req.Out.Host = upstream.Host
				req.Out.URL.Path = singleJoiningSlash(upstream.Path, strings.TrimPrefix(c.Request.URL.Path, "/"))
				req.Out.URL.RawPath = req.Out.URL.EscapedPath()
				req.Out.URL.RawQuery = c.Request.URL.RawQuery
				req.Out.Header.Set("X-Forwarded-Host", c.Request.Host)
				if c.Request.TLS != nil {
					req.Out.Header.Set("X-Forwarded-Proto", "https")
				} else {
					req.Out.Header.Set("X-Forwarded-Proto", "http")
				}
				req.Out.Header.Set("X-Forwarded-For", c.ClientIP())
			},
			FlushInterval: buildFlushInterval(h.useCase.FlushInterval()),
		}

		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
			var bandwidthErr *responseBandwidthLimitError
			if errors.As(err, &bandwidthErr) {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusTooManyRequests)
				_, _ = writer.Write([]byte(`{"error":"rate limit exceeded","type":"` + bandwidthErr.decision.LimitType + `"}`))
				return
			}

			upstreamLatency := time.Since(startedAt)
			c.Set("upstream_latency", upstreamLatency)
			h.monitoring.RecordUpstream(c.Request.Context(), usecase.RecordUpstreamInput{
				Path:       c.Request.URL.Path,
				StatusCode: http.StatusBadGateway,
				Latency:    upstreamLatency,
				Error:      err.Error(),
			})
			newReverseProxyErrorHandler()(writer, request, err)
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (h *ProxyController) serveCachedResponse(c *gin.Context, request domain.CacheRequest) (domain.CacheEntry, bool) {
	if h.cache == nil {
		c.Set(proxyContextCacheStatusKey, "bypass")
		return domain.CacheEntry{}, false
	}

	entry, exists, err := h.cache.Lookup(c.Request.Context(), request)
	if err != nil {
		c.Set(proxyContextCacheStatusKey, "lookup_error")
		return domain.CacheEntry{}, false
	}
	if !exists {
		c.Set(proxyContextCacheStatusKey, "miss")
		h.monitoring.RecordCacheEvent(c.Request.Context(), usecase.RecordCacheInput{
			Key:    "",
			Domain: request.Domain,
			Path:   request.Path,
			Action: "miss",
		})
		return domain.CacheEntry{}, false
	}

	downloadSize := int64(len(entry.Body))
	if c.Request.Method == http.MethodHead && entry.Metadata.Size > 0 {
		downloadSize = entry.Metadata.Size
	}
	if downloadSize > 0 {
		decision, err := h.rateLimit.ReserveResponseBandwidth(c.Request.Context(), c.ClientIP(), downloadSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return domain.CacheEntry{}, true
		}
		if !decision.Allowed {
			c.Set(proxyContextRateLimitDecisionKey, "deny")
			h.monitoring.RecordRateLimitDecision(c.Request.Context(), usecase.RecordRateLimitInput{
				IP:           decision.IP,
				RuleID:       decision.RuleID,
				LimitType:    decision.LimitType,
				Allowed:      decision.Allowed,
				CurrentValue: decision.CurrentValue,
			})
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"type":  decision.LimitType,
			})
			return domain.CacheEntry{}, true
		}
		c.Set(proxyContextReservedDownloadBytesKey, downloadSize)
	}

	c.Set(proxyContextCacheStatusKey, "hit")
	h.monitoring.RecordCacheEvent(c.Request.Context(), usecase.RecordCacheInput{
		Key:    entry.Metadata.Key,
		Domain: request.Domain,
		Path:   request.Path,
		Action: "hit",
	})
	writeCachedResponse(c.Writer, c.Request.Method, entry)
	return entry, true
}

func (h *ProxyController) captureAndStoreResponse(c *gin.Context, request domain.CacheRequest, response *http.Response) error {
	if h.cache == nil {
		c.Set(proxyContextCacheStatusKey, "disabled")
		return nil
	}

	size := response.ContentLength
	if response.Request != nil && response.Request.Method == http.MethodHead && size < 0 {
		size = 0
	}

	candidate := domain.CacheCandidate{
		Request:     request,
		StatusCode:  response.StatusCode,
		ContentType: response.Header.Get("Content-Type"),
		Headers:     sanitizeHeadersForCache(response.Header),
		Size:        size,
	}

	decision := h.cache.DecideCachePolicy(candidate)
	if !decision.Allowed {
		if contextCacheStatus(c) == "miss" {
			c.Set(proxyContextCacheStatusKey, "miss_bypass")
		}
		return nil
	}

	if response.Request != nil && response.Request.Method == http.MethodHead {
		if _, stored, err := h.cache.Set(c.Request.Context(), candidate, nil); err == nil && stored {
			c.Set(proxyContextCacheStatusKey, "miss_store")
			h.monitoring.RecordCacheEvent(c.Request.Context(), usecase.RecordCacheInput{
				Key:    decision.Key,
				Domain: request.Domain,
				Path:   request.Path,
				Action: "store",
				Rule:   decision.MatchedRule,
			})
		}
		return nil
	}

	if response.ContentLength < 0 {
		if contextCacheStatus(c) == "miss" {
			c.Set(proxyContextCacheStatusKey, "miss_bypass")
		}
		return nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	_ = response.Body.Close()

	candidate.Size = int64(len(body))
	decision = h.cache.DecideCachePolicy(candidate)
	if !decision.Allowed {
		response.Body = io.NopCloser(bytes.NewReader(body))
		response.ContentLength = int64(len(body))
		response.Header.Set("Content-Length", strconv.Itoa(len(body)))
		if contextCacheStatus(c) == "miss" {
			c.Set(proxyContextCacheStatusKey, "miss_bypass")
		}
		return nil
	}

	if _, stored, err := h.cache.Set(c.Request.Context(), candidate, body); err == nil && stored {
		c.Set(proxyContextCacheStatusKey, "miss_store")
		h.monitoring.RecordCacheEvent(c.Request.Context(), usecase.RecordCacheInput{
			Key:    decision.Key,
			Domain: request.Domain,
			Path:   request.Path,
			Action: "store",
			Rule:   decision.MatchedRule,
		})
	}

	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	response.Header.Set("Content-Length", strconv.Itoa(len(body)))
	return nil
}

func buildCacheRequest(request *http.Request) domain.CacheRequest {
	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}

	return domain.CacheRequest{
		Method:   request.Method,
		Scheme:   scheme,
		Domain:   canonicalHost(request.Host),
		Path:     request.URL.Path,
		RawQuery: request.URL.RawQuery,
		Headers:  cloneHeader(request.Header),
	}
}

func writeCachedResponse(writer http.ResponseWriter, method string, entry domain.CacheEntry) {
	headers := writer.Header()
	for key, values := range entry.Metadata.Headers {
		for _, value := range values {
			headers.Add(key, value)
		}
	}
	if entry.Metadata.ContentType != "" && headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", entry.Metadata.ContentType)
	}
	if method != http.MethodHead && len(entry.Body) > 0 {
		headers.Set("Content-Length", strconv.Itoa(len(entry.Body)))
	}
	writer.WriteHeader(entry.Metadata.StatusCode)
	if method != http.MethodHead {
		_, _ = writer.Write(entry.Body)
	}
}

func sanitizeHeadersForCache(header http.Header) map[string][]string {
	cloned := cloneHeader(header)
	removeHopByHopHeaders(cloned)
	return cloned
}

func cloneHeader(header http.Header) map[string][]string {
	cloned := make(map[string][]string, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func removeHopByHopHeaders(header map[string][]string) {
	connectionHeaders := header["Connection"]
	for _, value := range connectionHeaders {
		for _, token := range strings.Split(value, ",") {
			delete(header, http.CanonicalHeaderKey(strings.TrimSpace(token)))
		}
	}

	for _, key := range []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	} {
		delete(header, key)
	}
}

func contextCacheStatus(c *gin.Context) string {
	value, exists := c.Get(proxyContextCacheStatusKey)
	if !exists {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func canonicalHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	return strings.Trim(host, "[]")
}

func singleJoiningSlash(basePath, requestPath string) string {
	switch {
	case strings.HasSuffix(basePath, "/") && strings.HasPrefix(requestPath, "/"):
		return basePath + requestPath[1:]
	case !strings.HasSuffix(basePath, "/") && !strings.HasPrefix(requestPath, "/"):
		if requestPath == "" {
			return basePath
		}
		return basePath + "/" + requestPath
	default:
		return basePath + requestPath
	}
}
