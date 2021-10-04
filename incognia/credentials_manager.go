package incognia

import (
	"encoding/json"
	"net/http"
	"time"
)

type clientCredentialsTokenManager struct {
	ClientID     string
	clientSecret string
	netClient    *http.Client
	token        *accessToken
}

func newClientCredentialsTokenManager(ClientID, clientSecret string) *clientCredentialsTokenManager {
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	return &clientCredentialsTokenManager{
		ClientID,
		clientSecret,
		netClient,
		nil,
	}
}

func (tokenManager *clientCredentialsTokenManager) getToken() *accessToken {
	if tokenManager.token == nil || !tokenManager.token.isValid() {
		tokenManager.refreshToken()
	}

	return tokenManager.token
}

func (tokenManager *clientCredentialsTokenManager) refreshToken() error {
	req, _ := http.NewRequest("POST", tokenEndpoint, nil)

	req.SetBasicAuth(tokenManager.ClientID, tokenManager.clientSecret)
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
