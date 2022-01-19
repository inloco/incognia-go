package incognia

import (
	"errors"
	"time"
)

type TokenProvider interface {
	GetToken() (Token, error)
}

type Token interface {
	IsExpired() bool
	GetExpiresAt() time.Time
}

var (
	ErrTokenNotFound = errors.New("token not found in memory")
	ErrTokenExpired  = errors.New("incognia token expired")
)

type ManualRefreshTokenProvider struct {
	tokenClient *TokenClient
	token       Token
}

func NewManualRefreshTokenProvider(tokenClient *TokenClient) *ManualRefreshTokenProvider {
	return &ManualRefreshTokenProvider{tokenClient: tokenClient}
}

func (t *ManualRefreshTokenProvider) GetToken() (Token, error) {
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

	t.token = accessToken

	return t.token, nil
}

type AutoRefreshTokenProvider struct {
	tokenClient *TokenClient
	token       Token
}

func NewAutoRefreshTokenProvider(tokenClient *TokenClient) *AutoRefreshTokenProvider {
	return &AutoRefreshTokenProvider{
		tokenClient: tokenClient,
	}
}

func (t *AutoRefreshTokenProvider) GetToken() (Token, error) {
	if t.token == nil || t.token.IsExpired() {
		return t.refresh()
	}

	return t.token, nil
}

func (t *AutoRefreshTokenProvider) refresh() (Token, error) {
	accessToken, err := t.tokenClient.requestToken()
	if err != nil {
		return nil, err
	}

	t.token = accessToken

	return t.token, nil
}
