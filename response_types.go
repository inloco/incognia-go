package incognia

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type accessToken struct {
	CreatedAt   int64
	ExpiresIn   int64  `json:"expires_in,string"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (token accessToken) IsExpired() bool {
	expiresAt := token.GetExpiresAt()
	return time.Now().After(expiresAt)
}

func (token accessToken) GetExpiresAt() time.Time {
	createdAt := token.CreatedAt
	expiresIn := token.ExpiresIn
	return time.Unix(createdAt+expiresIn, 0)
}

func (token accessToken) Type() string {
	return token.TokenType
}

func (token accessToken) SetAuthHeader(request *http.Request) {
	request.Header.Add("Authorization", fmt.Sprintf("%s %s", token.Type(), token.AccessToken))
}

type Assessment string

const (
	LowRisk     Assessment = "low_risk"
	HighRisk    Assessment = "high_risk"
	UnknownRisk Assessment = "unknown_risk"
)

var (
	ErrEvidenceNotFound = errors.New("evidence not found")
)

type Evidence map[string]interface{}

type Reason struct {
	Code   string
	Source string
}

func (a Evidence) GetEvidence(evidenceName string, evidenceOut interface{}) error {
	return a.getEvidenceWithPath(a, strings.Split(evidenceName, "."), evidenceOut)
}

func (a Evidence) GetEvidenceAsInt64(evidenceName string) (int64, error) {
	var evidenceOut float64
	if err := a.GetEvidence(evidenceName, &evidenceOut); err != nil {
		return 0, err
	}

	for math.Mod(evidenceOut, 1) != 0 {
		evidenceOut *= 10
	}

	return int64(evidenceOut), nil
}

func (a Evidence) getEvidenceWithPath(evidenceMap Evidence, evidencePath []string, evidenceOut interface{}) error {
	if evidenceMap == nil {
		return ErrEvidenceNotFound
	}

	if len(evidencePath) == 0 {
		return ErrEvidenceNotFound
	}

	for len(evidencePath) > 1 {
		evidenceName := evidencePath[0]
		evidencePath = evidencePath[1:]

		evidence, ok := evidenceMap[evidenceName]
		if !ok {
			return ErrEvidenceNotFound
		}

		evidenceSubMap, ok := evidence.(map[string]interface{})
		if !ok {
			return ErrEvidenceNotFound
		}

		evidenceMap = evidenceSubMap
	}

	return a.getEvidence(evidenceMap, evidencePath[0], evidenceOut)
}

func (a Evidence) getEvidence(evidenceMap map[string]interface{}, evidenceName string, evidenceOut interface{}) error {
	if evidenceMap == nil {
		return ErrEvidenceNotFound
	}

	evidence, ok := evidenceMap[evidenceName]
	if !ok {
		return ErrEvidenceNotFound
	}
	if evidence == nil {
		return ErrEvidenceNotFound
	}

	if evidenceSlice, ok := evidence.([]interface{}); ok {
		return a.setEvidenceToSlice(evidenceSlice, evidenceOut)
	}

	return a.setEvidenceToPointer(evidence, evidenceOut)
}

func (a Evidence) setEvidenceToPointer(evidence interface{}, evidenceOut interface{}) error {
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

func (a Evidence) setEvidenceToSlice(evidenceSlice []interface{}, evidenceOut interface{}) error {
	evidenceOutReflectValue := reflect.ValueOf(evidenceOut)
	if evidenceOutReflectValue.Kind() != reflect.Ptr {
		return errors.New("expecting evidenceOut to be a pointer to slice")
	}
	evidenceOutIndirectReflectValue := reflect.Indirect(evidenceOutReflectValue)

	if evidenceOutIndirectReflectValue.Kind() != reflect.Slice {
		return errors.New("expecting evidenceOut to be a pointer to slice")
	}
	evidenceOutElemIndirectReflectKind := evidenceOutIndirectReflectValue.Type().Elem().Kind()

	if len(evidenceSlice) == 0 {
		return nil
	}
	evidenceSliceElemKind := reflect.ValueOf(evidenceSlice[0]).Kind()

	for _, e := range evidenceSlice[1:] {
		if reflect.ValueOf(e).Kind() != evidenceSliceElemKind {
			evidenceSliceElemKind = reflect.Interface
			break
		}
	}

	if evidenceOutElemIndirectReflectKind != evidenceSliceElemKind {
		return fmt.Errorf("expecting evidenceOut to be a pointer to slice of %s", evidenceSliceElemKind.String())
	}

	evidenceSliceReflectValue := reflect.MakeSlice(evidenceOutIndirectReflectValue.Type(), 0, len(evidenceSlice))
	for _, e := range evidenceSlice {
		evidenceSliceReflectValue = reflect.Append(evidenceSliceReflectValue, reflect.ValueOf(e))
	}

	evidenceOutIndirectReflectValue.Set(evidenceSliceReflectValue)
	return nil
}

type SignupAssessment struct {
	ID             string     `json:"id"`
	DeviceID       string     `json:"device_id"`
	RequestID      string     `json:"request_id"`
	RiskAssessment Assessment `json:"risk_assessment"`
	Evidence       Evidence   `json:"evidence"`
	Reasons        []Reason   `json:"reasons"`
}

type TransactionAssessment struct {
	ID             string     `json:"id"`
	RiskAssessment Assessment `json:"risk_assessment"`
	DeviceID       string     `json:"device_id"`
	Evidence       Evidence   `json:"evidence"`
	Reasons        []Reason   `json:"reasons"`
}
