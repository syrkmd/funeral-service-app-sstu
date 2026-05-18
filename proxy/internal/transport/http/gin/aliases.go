package gintransport

import "github.com/syrkmd/funeral-service-app-sstu/proxy/internal/middleware"

var (
	IPAccess      = middleware.IPAccess
	RateLimit     = middleware.RateLimit
	Recovery      = middleware.Recovery
	RequestLogger = middleware.RequestLogger
)
