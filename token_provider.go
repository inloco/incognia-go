package incognia

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found in memory")
	ErrTokenExpired  = errors.New("incognia token expired")
)

type TokenProvider interface {
	GetToken() (Token, error)
}

type Token interface {
	IsExpired() bool
	GetExpiresAt() time.Time
	Type() string
	SetAuthHeader(*http.Request)
}

type ManualRefreshTokenProvider struct {
	tokenClient *TokenClient
	token       Token
	tokenMutex  sync.RWMutex
}

func NewManualRefreshTokenProvider(tokenClient *TokenClient) *ManualRefreshTokenProvider {
	return &ManualRefreshTokenProvider{tokenClient: tokenClient}
}

func (t *ManualRefreshTokenProvider) GetToken() (Token, error) {
	t.tokenMutex.RLock()
	defer t.tokenMutex.RUnlock()

	if t.token == nil {
		return nil, ErrTokenNotFound
	}

	if t.token.IsExpired() {
		return nil, ErrTokenExpired
	}

	return t.token, nil
}

func (t *ManualRefreshTokenProvider) Refresh() (Token, error) {
	accessToken, err := t.tokenClient.requestToken()
	if err != nil {
		return nil, err
	}

	t.tokenMutex.Lock()
	defer t.tokenMutex.Unlock()

	t.token = accessToken

	return t.token, nil
}

type AutoRefreshTokenProvider struct {
	tokenClient *TokenClient
	token       Token
	tokenMutex  sync.RWMutex
}

func NewAutoRefreshTokenProvider(tokenClient *TokenClient) *AutoRefreshTokenProvider {
	return &AutoRefreshTokenProvider{
		tokenClient: tokenClient,
	}
}

func (t *AutoRefreshTokenProvider) GetToken() (Token, error) {
	var unlockOnce sync.Once

	t.tokenMutex.RLock()
	defer unlockOnce.Do(t.tokenMutex.RUnlock)

	if t.token == nil || t.token.IsExpired() {
		unlockOnce.Do(t.tokenMutex.RUnlock)
		return t.refresh()
	}

	return t.token, nil
}

func (t *AutoRefreshTokenProvider) refresh() (Token, error) {
	t.tokenMutex.Lock()
	defer t.tokenMutex.Unlock()

	if t.token != nil && !t.token.IsExpired() {
		return t.token, nil
	}

	accessToken, err := t.tokenClient.requestToken()
	if err != nil {
		return nil, err
	}

	t.token = accessToken

	return t.token, nil
}
