package gintransport

import "proxy/internal/middleware"

var (
	IPAccess      = middleware.IPAccess
	RateLimit     = middleware.RateLimit
	Recovery      = middleware.Recovery
	RequestLogger = middleware.RequestLogger
)
