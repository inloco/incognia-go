package incognia

import (
	"encoding/json"
	"net/http"
	"time"
)

type clientCredentialsTokenManager struct {
	ClientID      string
	ClientSecret  string
	NetClient     *http.Client
	Token         *accessToken
	TokenEndpoint string
}

type clientCredentialsManagerConfig struct {
	ClientID     string
	ClientSecret string
	Endpoint     string
	NetClient    *http.Client
}

func newClientCredentialsTokenManager(config *clientCredentialsManagerConfig) *clientCredentialsTokenManager {
	return &clientCredentialsTokenManager{
		ClientID:      config.ClientID,
		ClientSecret:  config.ClientSecret,
		NetClient:     config.NetClient,
		TokenEndpoint: config.Endpoint,
	}
}

func (tm *clientCredentialsTokenManager) getToken() (*accessToken, error) {
	if tm.Token == nil || !tm.Token.isValid() {
		err := tm.refreshToken()

		if err != nil {
			return nil, err
		}
	}

	return tm.Token, nil
}

func (tm *clientCredentialsTokenManager) refreshToken() error {
	req, err := http.NewRequest("POST", tm.TokenEndpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(tm.ClientID, tm.ClientSecret)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := tm.NetClient.Do(req)
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

	tm.Token = &result

	return nil
}
