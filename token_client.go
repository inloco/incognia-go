package incognia

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

const (
	tokenNetClientTimeout = 5 * time.Second
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type TokenClient struct {
	ClientID      string
	ClientSecret  string
	netClient     *http.Client
	tokenEndpoint string
}

type TokenClientConfig struct {
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

func NewTokenClient(config *TokenClientConfig) *TokenClient {
	incogniaEndpoints := getEndpoints()

	timeout := config.Timeout
	if timeout == 0 {
		timeout = tokenNetClientTimeout
	}

	return &TokenClient{
		ClientID:      config.ClientID,
		ClientSecret:  config.ClientSecret,
		netClient:     &http.Client{Timeout: timeout},
		tokenEndpoint: incogniaEndpoints.Token,
	}
}

func (tm TokenClient) requestToken() (Token, error) {
	req, err := http.NewRequest("POST", tm.tokenEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(tm.ClientID, tm.ClientSecret)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := tm.netClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidCredentials
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Error refreshing token: " + strconv.Itoa(res.StatusCode))
	}

	result := &accessToken{
		CreatedAt: time.Now().Unix(),
	}

	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}
