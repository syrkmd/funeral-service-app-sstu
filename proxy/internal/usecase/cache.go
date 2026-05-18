package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/config"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type compiledCacheConfig struct {
	GlobalRule    compiledCacheRule
	SpecificRules []compiledCacheRule
}

type compiledCacheRule struct {
	domain.CacheRule
	ExcludedHeaders      map[string]struct{}
	ExcludedContentTypes []string
	DomainMatchers       []compiledDomainMatcher
	PathMatchers         []compiledPathMatcher
	DomainPriority       int
	PathPriority         int
	PathLength           int
	Order                int
}

type compiledDomainMatcher struct {
	Pattern string
	Suffix  string
	Exact   bool
	Length  int
}

type compiledPathMatcher struct {
	Pattern string
	Prefix  string
	Exact   bool
	Length  int
}

type CacheUseCase struct {
	repo   CacheRepository
	config ConfigProvider

	compiled atomic.Value
}

func NewCacheUseCase(repo CacheRepository, provider ConfigProvider) *CacheUseCase {
	useCase := &CacheUseCase{
		repo:   repo,
		config: provider,
	}

	useCase.compiled.Store(compiledCacheConfig{})
	useCase.applyConfig(provider.Current())

	provider.Subscribe(func(cfg config.Config) {
		useCase.applyConfig(cfg)
	})

	return useCase
}

func (u *CacheUseCase) GenerateCacheKey(input domain.CacheRequest) string {
	canonical := strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(input.Method)),
		strings.ToLower(strings.TrimSpace(input.Scheme)),
		strings.ToLower(strings.TrimSpace(input.Domain)),
		strings.TrimSpace(input.Path),
		strings.TrimSpace(input.RawQuery),
	}, "|")

	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func (u *CacheUseCase) ShouldCache(input domain.CacheCandidate) bool {
	return u.DecideCachePolicy(input).Allowed
}

func (u *CacheUseCase) DecideCachePolicy(input domain.CacheCandidate) domain.CacheDecision {
	rule := u.resolveRule(input.Request)
	if decision, denied := evaluateRequestCacheRestrictions(input.Request, input.Headers, rule); denied {
		return decision
	}

	if isExcludedContentType(input.ContentType, rule.ExcludedContentTypes) {
		return domain.CacheDecision{Allowed: false, Reason: "content_type_excluded", MatchedRule: rule.Name}
	}

	if input.Size < 0 {
		return domain.CacheDecision{Allowed: false, Reason: "unknown_size", MatchedRule: rule.Name}
	}

	if input.Size < rule.MinSize {
		return domain.CacheDecision{Allowed: false, Reason: "entry_too_small", MatchedRule: rule.Name}
	}
	if rule.MaxSize > 0 && input.Size > rule.MaxSize {
		return domain.CacheDecision{Allowed: false, Reason: "entry_too_large", MatchedRule: rule.Name}
	}

	ttl, allowed := resolveTTL(rule.CacheRule, input.StatusCode)
	if !allowed {
		return domain.CacheDecision{Allowed: false, Reason: "status_not_cacheable", MatchedRule: rule.Name}
	}
	if ttl <= 0 {
		return domain.CacheDecision{Allowed: false, Reason: "ttl_not_configured", MatchedRule: rule.Name}
	}

	return domain.CacheDecision{
		Allowed:     true,
		Key:         u.GenerateCacheKey(input.Request),
		TTL:         ttl,
		Reason:      "cacheable",
		MatchedRule: rule.Name,
	}
}

func (u *CacheUseCase) Lookup(ctx context.Context, request domain.CacheRequest) (domain.CacheEntry, bool, error) {
	rule := u.resolveRule(request)
	if _, denied := evaluateRequestCacheRestrictions(request, nil, rule); denied {
		return domain.CacheEntry{}, false, nil
	}
	return u.repo.Get(ctx, u.GenerateCacheKey(request))
}

func (u *CacheUseCase) Get(ctx context.Context, key string) (domain.CacheEntry, bool, error) {
	return u.repo.Get(ctx, key)
}

func (u *CacheUseCase) Set(ctx context.Context, candidate domain.CacheCandidate, body []byte) (domain.CacheEntry, bool, error) {
	decision := u.DecideCachePolicy(candidate)
	if !decision.Allowed {
		return domain.CacheEntry{}, false, nil
	}

	now := time.Now().UTC()
	entry := domain.CacheEntry{
		Metadata: domain.CacheMetadata{
			Key:         decision.Key,
			Headers:     cloneHeaders(candidate.Headers),
			StatusCode:  candidate.StatusCode,
			ContentType: candidate.ContentType,
			TTL:         decision.TTL,
			CreatedAt:   now,
			ExpiresAt:   now.Add(decision.TTL),
			Tags:        append([]string(nil), candidate.Tags...),
			Domain:      candidate.Request.Domain,
			Path:        candidate.Request.Path,
			Size:        candidate.Size,
		},
		Body: append([]byte(nil), body...),
	}

	if err := u.repo.Set(ctx, entry); err != nil {
		return domain.CacheEntry{}, false, err
	}

	return entry, true, nil
}

func (u *CacheUseCase) Delete(ctx context.Context, key string) error {
	_, err := u.repo.Delete(ctx, key)
	return err
}

func (u *CacheUseCase) InvalidateByKey(ctx context.Context, key string) (int, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, nil
	}
	return u.repo.Delete(ctx, key)
}

func (u *CacheUseCase) InvalidateByPrefix(ctx context.Context, prefix string) (int, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return 0, nil
	}
	return u.repo.DeleteByPrefix(ctx, prefix)
}

func (u *CacheUseCase) InvalidateByRegex(ctx context.Context, pattern string) (int, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return 0, nil
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return 0, err
	}

	return u.repo.DeleteByRegex(ctx, compiled)
}

func (u *CacheUseCase) InvalidateByTags(ctx context.Context, tags []string) (int, error) {
	normalized := normalizeStringList(tags)
	if len(normalized) == 0 {
		return 0, nil
	}
	return u.repo.DeleteByTags(ctx, normalized)
}

func (u *CacheUseCase) InvalidateExpired(ctx context.Context) (int, error) {
	return u.repo.CleanupExpired(ctx)
}

func (u *CacheUseCase) Clear(ctx context.Context) (int, error) {
	return u.repo.Clear(ctx)
}

func (u *CacheUseCase) HandleInvalidationEvent(ctx context.Context, event domain.CacheInvalidationEvent) (int, error) {
	deleted := 0

	for _, key := range normalizeStringList(event.Keys) {
		count, err := u.repo.Delete(ctx, key)
		if err != nil {
			return deleted, err
		}
		deleted += count
	}

	for _, prefix := range normalizeStringList(event.Prefixes) {
		count, err := u.repo.DeleteByPrefix(ctx, prefix)
		if err != nil {
			return deleted, err
		}
		deleted += count
	}

	if len(event.Tags) > 0 {
		count, err := u.repo.DeleteByTags(ctx, normalizeStringList(event.Tags))
		if err != nil {
			return deleted, err
		}
		deleted += count
	}

	if event.Regex != nil {
		count, err := u.repo.DeleteByRegex(ctx, event.Regex)
		if err != nil {
			return deleted, err
		}
		deleted += count
	}

	return deleted, nil
}

func (u *CacheUseCase) applyConfig(cfg config.Config) {
	cacheCfg := cfg.Cache
	globalRule := domain.CacheRule{
		Name:                 "global-default",
		Enabled:              cacheCfg.Enabled,
		DefaultTTL:           cacheCfg.DefaultTTL.Duration,
		TTL2XX:               cacheCfg.TTL2XX.Duration,
		TTL3XX:               cacheCfg.TTL3XX.Duration,
		TTL4XX:               cacheCfg.TTL4XX.Duration,
		TTL5XX:               cacheCfg.TTL5XX.Duration,
		MinSize:              cacheCfg.MinEntrySize,
		MaxSize:              cacheCfg.MaxEntrySize,
		Cache2XX:             cacheCfg.Cache2XX,
		Cache3XX:             cacheCfg.Cache3XX,
		Cache4XX:             cacheCfg.Cache4XX,
		Cache5XX:             cacheCfg.Cache5XX,
		ExcludedHeaders:      append([]string(nil), cacheCfg.ExcludedHeaders...),
		ExcludedContentTypes: append([]string(nil), cacheCfg.ExcludedContentTypes...),
	}
	compiled := compiledCacheConfig{
		GlobalRule:    compileCacheRule(globalRule, 0),
		SpecificRules: make([]compiledCacheRule, 0, len(cacheCfg.Rules)),
	}

	for idx, ruleCfg := range cacheCfg.Rules {
		rule := mergeCacheRule(globalRule, ruleCfg, idx+1)
		compiled.SpecificRules = append(compiled.SpecificRules, compileCacheRule(rule, idx+1))
	}

	sort.SliceStable(compiled.SpecificRules, func(i, j int) bool {
		left := compiled.SpecificRules[i]
		right := compiled.SpecificRules[j]
		if left.DomainPriority != right.DomainPriority {
			return left.DomainPriority > right.DomainPriority
		}
		if left.PathPriority != right.PathPriority {
			return left.PathPriority > right.PathPriority
		}
		if left.PathLength != right.PathLength {
			return left.PathLength > right.PathLength
		}
		return left.Order < right.Order
	})

	u.compiled.Store(compiled)
}

func (u *CacheUseCase) resolveRule(request domain.CacheRequest) compiledCacheRule {
	compiled := u.compiled.Load().(compiledCacheConfig)
	rule := compiled.GlobalRule
	for _, candidateRule := range compiled.SpecificRules {
		if candidateRule.matches(request) {
			rule = candidateRule
			break
		}
	}
	return rule
}

func evaluateRequestCacheRestrictions(request domain.CacheRequest, responseHeaders map[string][]string, rule compiledCacheRule) (domain.CacheDecision, bool) {
	if !rule.Enabled {
		return domain.CacheDecision{Allowed: false, Reason: "cache_disabled", MatchedRule: rule.Name}, true
	}

	method := strings.ToUpper(strings.TrimSpace(request.Method))
	if method != "GET" && method != "HEAD" {
		return domain.CacheDecision{Allowed: false, Reason: "method_not_cacheable", MatchedRule: rule.Name}, true
	}

	if hasSensitiveHeaders(request.Headers, rule.ExcludedHeaders) || hasSensitiveHeaders(responseHeaders, rule.ExcludedHeaders) {
		return domain.CacheDecision{Allowed: false, Reason: "sensitive_headers", MatchedRule: rule.Name}, true
	}

	if hasNoStoreDirectives(request.Headers) || hasNoStoreDirectives(responseHeaders) {
		return domain.CacheDecision{Allowed: false, Reason: "cache_control_restricted", MatchedRule: rule.Name}, true
	}

	return domain.CacheDecision{}, false
}

func resolveTTL(rule domain.CacheRule, statusCode int) (time.Duration, bool) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return firstPositiveDuration(rule.TTL2XX, rule.DefaultTTL), rule.Cache2XX
	case statusCode >= 300 && statusCode < 400:
		return firstPositiveDuration(rule.TTL3XX, rule.DefaultTTL), rule.Cache3XX
	case statusCode >= 400 && statusCode < 500:
		return firstPositiveDuration(rule.TTL4XX, rule.DefaultTTL), rule.Cache4XX
	case statusCode >= 500 && statusCode < 600:
		return firstPositiveDuration(rule.TTL5XX, rule.DefaultTTL), rule.Cache5XX
	default:
		return 0, false
	}
}

func firstPositiveDuration(values ...time.Duration) time.Duration {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func hasSensitiveHeaders(headers map[string][]string, excluded map[string]struct{}) bool {
	for key, values := range headers {
		if _, exists := excluded[strings.ToLower(key)]; exists && len(values) > 0 {
			return true
		}
	}
	return false
}

func hasNoStoreDirectives(headers map[string][]string) bool {
	for key, values := range headers {
		if !strings.EqualFold(key, "Cache-Control") {
			continue
		}
		for _, value := range values {
			normalized := strings.ToLower(value)
			if strings.Contains(normalized, "no-store") || strings.Contains(normalized, "private") || strings.Contains(normalized, "no-cache") {
				return true
			}
		}
	}
	return false
}

func isExcludedContentType(contentType string, excluded []string) bool {
	normalized := strings.ToLower(strings.TrimSpace(contentType))
	for _, candidate := range excluded {
		if strings.HasPrefix(normalized, candidate) {
			return true
		}
	}
	return false
}

func normalizeHeaderNames(headers []string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(headers))
	for _, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		if header == "" {
			continue
		}
		normalized[header] = struct{}{}
	}
	return normalized
}

func normalizeContentTypes(contentTypes []string) []string {
	normalized := make([]string, 0, len(contentTypes))
	for _, contentType := range contentTypes {
		contentType = strings.ToLower(strings.TrimSpace(contentType))
		if contentType == "" {
			continue
		}
		normalized = append(normalized, contentType)
	}
	return normalized
}

func compileCacheRule(rule domain.CacheRule, order int) compiledCacheRule {
	domainMatchers, domainPriority := compileDomainMatchers(rule.Domains)
	pathMatchers, pathPriority, pathLength := compilePathMatchers(rule.Paths)

	return compiledCacheRule{
		CacheRule:            rule,
		ExcludedHeaders:      normalizeHeaderNames(rule.ExcludedHeaders),
		ExcludedContentTypes: normalizeContentTypes(rule.ExcludedContentTypes),
		DomainMatchers:       domainMatchers,
		PathMatchers:         pathMatchers,
		DomainPriority:       domainPriority,
		PathPriority:         pathPriority,
		PathLength:           pathLength,
		Order:                order,
	}
}

func mergeCacheRule(global domain.CacheRule, ruleCfg config.CacheRuleConfig, index int) domain.CacheRule {
	enabled := global.Enabled
	if ruleCfg.Enabled != nil {
		enabled = *ruleCfg.Enabled
	}

	return domain.CacheRule{
		Name:                 defaultRuleName(ruleCfg.Name, index),
		Enabled:              enabled,
		Domains:              append([]string(nil), ruleCfg.Domains...),
		Paths:                append([]string(nil), ruleCfg.Paths...),
		DefaultTTL:           firstPositiveDuration(ruleCfg.DefaultTTL.Duration, global.DefaultTTL),
		TTL2XX:               firstPositiveDuration(ruleCfg.TTL2XX.Duration, global.TTL2XX),
		TTL3XX:               firstPositiveDuration(ruleCfg.TTL3XX.Duration, global.TTL3XX),
		TTL4XX:               firstPositiveDuration(ruleCfg.TTL4XX.Duration, global.TTL4XX),
		TTL5XX:               firstPositiveDuration(ruleCfg.TTL5XX.Duration, global.TTL5XX),
		MinSize:              selectInt64Ptr(ruleCfg.MinEntrySize, global.MinSize),
		MaxSize:              selectInt64Ptr(ruleCfg.MaxEntrySize, global.MaxSize),
		Cache2XX:             selectBool(ruleCfg.Cache2XX, global.Cache2XX),
		Cache3XX:             selectBool(ruleCfg.Cache3XX, global.Cache3XX),
		Cache4XX:             selectBool(ruleCfg.Cache4XX, global.Cache4XX),
		Cache5XX:             selectBool(ruleCfg.Cache5XX, global.Cache5XX),
		ExcludedHeaders:      mergeStringSlices(global.ExcludedHeaders, ruleCfg.ExcludedHeaders),
		ExcludedContentTypes: mergeStringSlices(global.ExcludedContentTypes, ruleCfg.ExcludedContentTypes),
		Tags:                 append([]string(nil), ruleCfg.Tags...),
	}
}

func (r compiledCacheRule) matches(request domain.CacheRequest) bool {
	return matchesDomain(strings.ToLower(strings.TrimSpace(request.Domain)), r.DomainMatchers) &&
		matchesPath(cleanPath(request.Path), r.PathMatchers)
}

func compileDomainMatchers(patterns []string) ([]compiledDomainMatcher, int) {
	matchers := make([]compiledDomainMatcher, 0, len(patterns))
	priority := 0
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if strings.HasPrefix(pattern, "*.") {
			suffix := strings.TrimPrefix(pattern, "*.")
			matchers = append(matchers, compiledDomainMatcher{
				Pattern: pattern,
				Suffix:  "." + suffix,
				Length:  len(suffix),
			})
			if priority < 1 {
				priority = 1
			}
			continue
		}
		matchers = append(matchers, compiledDomainMatcher{
			Pattern: pattern,
			Exact:   true,
			Length:  len(pattern),
		})
		if priority < 2 {
			priority = 2
		}
	}
	return matchers, priority
}

func compilePathMatchers(patterns []string) ([]compiledPathMatcher, int, int) {
	matchers := make([]compiledPathMatcher, 0, len(patterns))
	priority := 0
	maxLength := 0
	for _, pattern := range patterns {
		pattern = cleanPath(pattern)
		if pattern == "" {
			continue
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := cleanPath(strings.TrimSuffix(pattern, "*"))
			matchers = append(matchers, compiledPathMatcher{
				Pattern: pattern,
				Prefix:  prefix,
				Length:  len(prefix),
			})
			if priority < 1 {
				priority = 1
			}
			if len(prefix) > maxLength {
				maxLength = len(prefix)
			}
			continue
		}
		matchers = append(matchers, compiledPathMatcher{
			Pattern: pattern,
			Exact:   true,
			Length:  len(pattern),
		})
		if priority < 2 {
			priority = 2
		}
		if len(pattern) > maxLength {
			maxLength = len(pattern)
		}
	}
	return matchers, priority, maxLength
}

func matchesDomain(domainName string, matchers []compiledDomainMatcher) bool {
	if len(matchers) == 0 {
		return true
	}
	for _, matcher := range matchers {
		if matcher.Exact && domainName == matcher.Pattern {
			return true
		}
		if !matcher.Exact && strings.HasSuffix(domainName, matcher.Suffix) {
			return true
		}
	}
	return false
}

func matchesPath(pathValue string, matchers []compiledPathMatcher) bool {
	if len(matchers) == 0 {
		return true
	}
	for _, matcher := range matchers {
		if matcher.Exact && pathValue == matcher.Pattern {
			return true
		}
		if !matcher.Exact && strings.HasPrefix(pathValue, matcher.Prefix) {
			return true
		}
	}
	return false
}

func cleanPath(pathValue string) string {
	pathValue = strings.TrimSpace(pathValue)
	if pathValue == "" {
		return "/"
	}
	if !strings.HasPrefix(pathValue, "/") {
		pathValue = "/" + pathValue
	}
	if strings.HasSuffix(pathValue, "*") {
		prefix := strings.TrimSuffix(pathValue, "*")
		cleaned := path.Clean(prefix)
		if cleaned == "." {
			cleaned = "/"
		}
		if !strings.HasSuffix(cleaned, "/") {
			cleaned += "/"
		}
		return cleaned + "*"
	}
	cleaned := path.Clean(pathValue)
	if cleaned == "." {
		return "/"
	}
	return cleaned
}

func defaultRuleName(name string, index int) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	return "rule-" + strconv.Itoa(index)
}

func selectBool(value *bool, fallback bool) bool {
	if value != nil {
		return *value
	}
	return fallback
}

func selectInt64Ptr(value *int64, fallback int64) int64 {
	if value != nil {
		return *value
	}
	return fallback
}

func mergeStringSlices(base []string, override []string) []string {
	if len(override) == 0 {
		return append([]string(nil), base...)
	}
	return append([]string(nil), override...)
}

func cloneHeaders(headers map[string][]string) map[string][]string {
	cloned := make(map[string][]string, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}
