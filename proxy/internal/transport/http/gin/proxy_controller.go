package gintransport

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type ControllerProxyUseCase interface {
	ResolveUpstream(ctx context.Context) (*url.URL, error)
	FlushInterval() int64
}

type ProxyController struct {
	useCase ControllerProxyUseCase
}

func NewProxyController(useCase ControllerProxyUseCase) *ProxyController {
	return &ProxyController{useCase: useCase}
}

func (h *ProxyController) Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		upstream, err := h.useCase.ResolveUpstream(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		proxy := &httputil.ReverseProxy{
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

		withErrorHandler(proxy).ServeHTTP(c.Writer, c.Request)
	}
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
