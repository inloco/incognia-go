package incognia

import (
	"context"
	"errors"
	"time"
)

// PeriodicRefreshTokenProviderConfig controls proactive token refresh.
// Zero-valued fields use their defaults. Negative durations are invalid.
type PeriodicRefreshTokenProviderConfig struct {
	// RefreshBefore is the lead time before expiration. Default: one minute.
	// Must be less than the token lifetime of 15 minutes.
	RefreshBefore time.Duration
	// RetryInterval is the delay after a failed refresh. Default: five seconds.
	RetryInterval time.Duration
}

// PeriodicRefreshTokenProvider refreshes tokens in the background. GetToken
// only reads the cache and never performs network I/O. It must not be copied.
// Call Close when the provider is no longer needed.
type PeriodicRefreshTokenProvider struct {
	manual *ManualRefreshTokenProvider
	config PeriodicRefreshTokenProviderConfig
	cancel context.CancelFunc
	done   chan struct{}
}

var _ TokenProvider = (*PeriodicRefreshTokenProvider)(nil)

// NewPeriodicRefreshTokenProvider synchronously fetches the initial token, then
// starts a background refresh goroutine using the default configuration.
// If the initial fetch fails, no goroutine is started.
func NewPeriodicRefreshTokenProvider(tokenClient *TokenClient) (*PeriodicRefreshTokenProvider, error) {
	return NewPeriodicRefreshTokenProviderWithConfig(tokenClient, PeriodicRefreshTokenProviderConfig{})
}

// NewPeriodicRefreshTokenProviderWithConfig synchronously fetches the initial
// token, then starts a background refresh goroutine using config.
// If the initial fetch fails or config is invalid, no goroutine is started.
func NewPeriodicRefreshTokenProviderWithConfig(tokenClient *TokenClient, config PeriodicRefreshTokenProviderConfig) (*PeriodicRefreshTokenProvider, error) {
	if tokenClient == nil {
		return nil, errors.New("token client must not be nil")
	}
	if config.RefreshBefore < 0 || config.RetryInterval < 0 {
		return nil, errors.New("periodic refresh durations must not be negative")
	}
	if config.RefreshBefore >= 15*time.Minute {
		return nil, errors.New("refresh lead time must be less than 15 minutes")
	}
	if config.RefreshBefore == 0 {
		config.RefreshBefore = time.Minute
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = 5 * time.Second
	}

	manual := NewManualRefreshTokenProvider(tokenClient)
	token, err := manual.Refresh()
	if err != nil {
		return nil, err
	}
	if token.IsExpired() {
		return nil, ErrTokenExpired
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &PeriodicRefreshTokenProvider{
		manual: manual,
		config: config,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go p.run(ctx, token)
	return p, nil
}

// GetToken returns the cached token, or ErrTokenExpired if it has expired.
// Failed background refreshes retain the cached token and are retried silently.
func (p *PeriodicRefreshTokenProvider) GetToken() (Token, error) {
	return p.manual.GetToken()
}

// Close cancels any in-flight background refresh and waits for the goroutine
// to exit. It is safe to call concurrently or repeatedly. Cached tokens remain
// available through GetToken until they expire.
func (p *PeriodicRefreshTokenProvider) Close() error {
	p.cancel()
	<-p.done
	return nil
}

func (p *PeriodicRefreshTokenProvider) refreshDelay(token Token, now time.Time) time.Duration {
	return token.GetExpiresAt().Sub(now) - p.config.RefreshBefore
}

func (p *PeriodicRefreshTokenProvider) run(ctx context.Context, token Token) {
	defer close(p.done)
	timer := time.NewTimer(p.refreshDelay(token, time.Now()))
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		if ctx.Err() != nil {
			return
		}
		token, err := p.manual.refreshWithContext(ctx)
		delay := p.config.RetryInterval
		if err == nil {
			delay = p.refreshDelay(token, time.Now())
		}
		timer.Reset(delay)
	}
}
