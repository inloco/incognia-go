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
	ErrSignalNotFound   = errors.New("signal not found")
)

type Evidence map[string]interface{}

type Signals map[string]interface{}

type Reason struct {
	Code   string
	Source string
}

type jsonMap map[string]interface{}

func (a Evidence) GetEvidence(evidenceName string, outValue interface{}) error {
	if a == nil {
		return ErrEvidenceNotFound
	}
	return getValueWithPath(jsonMap(a), evidenceName, outValue)
}

func (a Evidence) GetEvidenceAsInt64(evidenceName string) (int64, error) {
	if a == nil {
		return 0, ErrEvidenceNotFound
	}

	var outValue float64
	if err := a.GetEvidence(evidenceName, &outValue); err != nil {
		return 0, err
	}

	for math.Mod(outValue, 1) != 0 {
		outValue *= 10
	}

	return int64(outValue), nil
}

func (s Signals) GetSignal(signalName string, outValue interface{}) error {
	if s == nil {
		return ErrSignalNotFound
	}
	return getValueWithPath(jsonMap(s), signalName, outValue)
}

func (s Signals) GetSignalAsInt64(signalName string) (int64, error) {
	if s == nil {
		return 0, ErrSignalNotFound
	}

	var outValue float64
	if err := s.GetSignal(signalName, &outValue); err != nil {
		return 0, err
	}

	for math.Mod(outValue, 1) != 0 {
		outValue *= 10
	}

	return int64(outValue), nil
}

func getValueWithPath(root jsonMap, path string, outValue interface{}) error {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return ErrEvidenceNotFound
	}

	curr := map[string]interface{}(root)
	for len(parts) > 1 {
		key := parts[0]
		parts = parts[1:]

		v, ok := curr[key]
		if !ok || v == nil {
			return ErrEvidenceNotFound
		}

		next, ok := v.(map[string]interface{})
		if !ok || next == nil {
			return ErrEvidenceNotFound
		}
		curr = next
	}

	lastKey := parts[0]
	v, ok := curr[lastKey]
	if !ok || v == nil {
		return ErrEvidenceNotFound
	}

	if slice, ok := v.([]interface{}); ok {
		return setToSlice(slice, outValue)
	}
	return setToPointer(v, outValue)
}

func setToPointer(value interface{}, outValue interface{}) error {
	outputReflectValue := reflect.ValueOf(outValue)
	if outputReflectValue.Kind() != reflect.Ptr {
		return errors.New("expecting outValue to be a pointer")
	}
	indirectOutputValueKind := reflect.Indirect(outputReflectValue).Kind()

	valueReflectValue := reflect.ValueOf(value)
	valueReflectKind := valueReflectValue.Kind()

	if indirectOutputValueKind != valueReflectKind {
		return fmt.Errorf("expecting outValue to be a pointer to %s", valueReflectKind.String())
	}

	outputReflectValue.Elem().Set(valueReflectValue)
	return nil
}

func setToSlice(slice []interface{}, outValue interface{}) error {
	outputReflectValue := reflect.ValueOf(outValue)
	if outputReflectValue.Kind() != reflect.Ptr {
		return errors.New("expecting outValue to be a pointer to slice")
	}
	indirectOutputValue := reflect.Indirect(outputReflectValue)

	if indirectOutputValue.Kind() != reflect.Slice {
		return errors.New("expecting outValue to be a pointer to slice")
	}
	indirectValueElementKind := indirectOutputValue.Type().Elem().Kind()

	if len(slice) == 0 {
		return nil
	}

	sliceElemKind := reflect.ValueOf(slice[0]).Kind()
	for _, e := range slice[1:] {
		if reflect.ValueOf(e).Kind() != sliceElemKind {
			sliceElemKind = reflect.Interface
			break
		}
	}

	if indirectValueElementKind != sliceElemKind {
		return fmt.Errorf("expecting outValue to be a pointer to slice of %s", sliceElemKind.String())
	}

	sliceReflectValue := reflect.MakeSlice(indirectOutputValue.Type(), 0, len(slice))
	for _, e := range slice {
		sliceReflectValue = reflect.Append(sliceReflectValue, reflect.ValueOf(e))
	}

	indirectOutputValue.Set(sliceReflectValue)
	return nil
}

type SignupAssessment struct {
	ID             string     `json:"id"`
	DeviceID       string     `json:"device_id"`
	RequestID      string     `json:"request_id"`
	RiskAssessment Assessment `json:"risk_assessment"`
	Evidence       Evidence   `json:"evidence,omitempty"`
	Reasons        []Reason   `json:"reasons"`
	Actions        []string   `json:"actions,omitempty"`
	Signals        Signals    `json:"signals,omitempty"`
}

type TransactionAssessment struct {
	ID             string     `json:"id"`
	RiskAssessment Assessment `json:"risk_assessment"`
	DeviceID       string     `json:"device_id"`
	Evidence       Evidence   `json:"evidence,omitempty"`
	Reasons        []Reason   `json:"reasons"`
	Actions        []string   `json:"actions,omitempty"`
	Signals        Signals    `json:"signals,omitempty"`
}
