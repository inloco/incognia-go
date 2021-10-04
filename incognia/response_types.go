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

type Assessment string

const (
	LowRisk     Assessment = "low_risk"
	HighRisk    Assessment = "high_risk"
	UnknownRisk Assessment = "unknown_risk"
)

type SignupAssessment struct {
	ID             string                 `json:"id"`
	DeviceID       string                 `json:"device_id"`
	RequestID      string                 `json:"request_id"`
	RiskAssessment Assessment             `json:"risk_assessment"`
	Evidence       map[string]interface{} `json:"evidence"`
}

type TransactionAssessment struct {
	Id             string                 `json:"id"`
	RiskAssessment Assessment             `json:"risk_assessment"`
	DeviceID       string                 `json:"device_id"`
	Evidence       map[string]interface{} `json:"evidence"`
}
