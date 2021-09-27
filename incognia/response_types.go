package incognia

import "time"

type accessToken struct {
	CreatedAt   int64
	ExpiresIn   int64  `json:"expires_in,string"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (token *accessToken) isValid() bool {
	createdAt := token.CreatedAt
	expiresIn := token.ExpiresIn

	expirationLimit := createdAt + expiresIn
	nowInSeconds := time.Now().Unix()

	if nowInSeconds >= expirationLimit {
		return false
	}

	return true
}
