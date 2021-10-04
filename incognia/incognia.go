package incognia

import (
	"errors"
	"net/http"
	"time"
)

type Client struct {
	clientId     string
	clientSecret string
	tokenManager *clientCredentialsTokenManager
	netClient    *http.Client
}

type IncogniaClientConfig struct {
	ClientId     string
	ClientSecret string
}

func New(config *IncogniaClientConfig) (*Client, error) {
	if config.ClientId == "" || config.ClientSecret == "" {
		return nil, errors.New("client id and client secret are required")
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	tokenManager := newClientCredentialsTokenManager(config.ClientId, config.ClientSecret)

	client := &Client{config.ClientId, config.ClientSecret, tokenManager, netClient}

	return client, nil
}
