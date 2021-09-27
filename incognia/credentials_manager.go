package incognia

import (
	"encoding/json"
	"net/http"
	"time"
)

type clientCredentialsTokenManager struct {
	clientId      string
	clientSecret  string
	tokenEndpoint string
	netClient     *http.Client
	token         *accessToken
}

func newClientCredentialsTokenManager(clientId, clientSecret, tokenEndpoint string) *clientCredentialsTokenManager {
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	return &clientCredentialsTokenManager{
		clientId,
		clientSecret,
		tokenEndpoint,
		netClient,
		nil,
	}
}

func (tokenManager *clientCredentialsTokenManager) getToken() *accessToken {
	if !tokenManager.token.isValid() {
		tokenManager.refreshToken()
	}

	return tokenManager.token
}

func (tokenManager *clientCredentialsTokenManager) refreshToken() error {
	req, _ := http.NewRequest("POST", tokenManager.tokenEndpoint, nil)

	req.SetBasicAuth(tokenManager.clientId, tokenManager.clientSecret)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := tokenManager.netClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var result accessToken

	result.CreatedAt = time.Now().Unix()

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return err
	}

	tokenManager.token = &result

	return nil
}
