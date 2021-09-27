package incognia

import (
	"errors"
	"net/http"
	"time"
)

type Client struct {
	clientId     string
	clientSecret string
	endpoints    endpoints
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

	endpoints := newEndpoints()
	tokenManager := newClientCredentialsTokenManager(config.ClientId, config.ClientSecret, endpoints.Token)

	client := &Client{config.ClientId, config.ClientSecret, endpoints, tokenManager, netClient}

	return client, nil
}
