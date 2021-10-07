package incognia

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Assessment string

const (
	LowRisk     Assessment = "low_risk"
	HighRisk    Assessment = "high_risk"
	UnknownRisk Assessment = "unknown_risk"
)

var (
	ErrEvidenceNotFoundError = errors.New("evidence not found")
)

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

type SignupAssessment struct {
	ID             string                 `json:"id"`
	DeviceID       string                 `json:"device_id"`
	RequestID      string                 `json:"request_id"`
	RiskAssessment Assessment             `json:"risk_assessment"`
	evidence       map[string]interface{} `json:"evidence"`
}

func (a *SignupAssessment) GetEvidence(evidenceName string, evidenceOut interface{}) error {
	evidence, err := getEvidenceValue(a.evidence, evidenceName)
	if err != nil {
		return err
	}

	return readEvidence(evidence, evidenceOut)
}

func (a *SignupAssessment) GetEvidenceAsInt64(evidenceName string) (int64, error) {
	evidence, err := getEvidenceValue(a.evidence, evidenceName)
	if err != nil {
		return 0, err
	}

	var evidenceValue float64
	readEvidence(evidence, evidenceValue)

	return int64(evidenceValue), nil
}

func (a *SignupAssessment) GetEvidenceAsTime(evidenceName string) (time.Time, error) {
	evidence, err := a.GetEvidenceAsInt64(evidenceName)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, evidence*int64(1000000)), nil
}

type TransactionAssessment struct {
	ID             string                 `json:"id"`
	RiskAssessment Assessment             `json:"risk_assessment"`
	DeviceID       string                 `json:"device_id"`
	evidence       map[string]interface{} `json:"evidence"`
}

func (a *TransactionAssessment) GetEvidence(evidenceName string, evidenceOut interface{}) error {
	evidence, err := getEvidenceValue(a.evidence, evidenceName)
	if err != nil {
		return err
	}

	return readEvidence(evidence, evidenceOut)
}

func (a *TransactionAssessment) GetEvidenceAsInt64(evidenceName string) (int64, error) {
	evidence, err := getEvidenceValue(a.evidence, evidenceName)
	if err != nil {
		return 0, err
	}

	var evidenceValue float64
	readEvidence(evidence, evidenceValue)

	return int64(evidenceValue), nil
}

func (a *TransactionAssessment) GetEvidenceAsTime(evidenceName string) (time.Time, error) {
	evidence, err := a.GetEvidenceAsInt64(evidenceName)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, evidence*int64(1000000)), nil
}

func getEvidenceValue(evidenceMap map[string]interface{}, evidenceName string) (interface{}, error) {
	if evidenceMap == nil {
		return nil, ErrEvidenceNotFoundError
	}

	evidence, ok := evidenceMap[evidenceName]
	if !ok {
		return 0, ErrEvidenceNotFoundError
	}

	return evidence, nil
}

func readEvidence(evidence interface{}, evidenceOut interface{}) error {
	evidenceOutReflectValue := reflect.ValueOf(evidenceOut)
	if evidenceOutReflectValue.Kind() != reflect.Ptr {
		return errors.New("expecting evidenceOut to be a pointer")
	}
	evidenceOutIndirectReflectKind := reflect.Indirect(evidenceOutReflectValue).Kind()

	evidenceReflectValue := reflect.ValueOf(evidence)
	evidenceReflectKind := evidenceReflectValue.Kind()

	if evidenceOutIndirectReflectKind != evidenceReflectKind {
		return fmt.Errorf("expecting evidenceOut to be a pointer to %s", evidenceReflectKind.String())
	}

	evidenceOutReflectValue.Elem().Set(evidenceReflectValue)
	return nil
}
