package usecase

import (
	"context"
	"net/url"

	"proxy/internal/domain"
)

type ProxyUseCase struct {
	config ConfigProvider
}

func NewProxyUseCase(config ConfigProvider) *ProxyUseCase {
	return &ProxyUseCase{config: config}
}

func (u *ProxyUseCase) ResolveUpstream(_ context.Context) (*url.URL, error) {
	upstreamURL, err := url.Parse(u.config.Current().Proxy.UpstreamURL)
	if err != nil || upstreamURL.Scheme == "" || upstreamURL.Host == "" {
		return nil, domain.ErrInvalidUpstream
	}

	return upstreamURL, nil
}

func (u *ProxyUseCase) FlushInterval() int64 {
	return int64(u.config.Current().Proxy.FlushInterval.Duration)
}
