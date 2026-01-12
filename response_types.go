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

type Signals map[string]interface{}

type Reason struct {
	Code   string
	Source string
}

// jsonMap is an internal helper type to share "get by path" logic between Evidence and Signals.
type jsonMap map[string]interface{}

func (a Evidence) GetEvidence(evidenceName string, evidenceOut interface{}) error {
	if a == nil {
		return ErrEvidenceNotFound
	}
	return getValueWithPath(jsonMap(a), evidenceName, evidenceOut)
}

func (a Evidence) GetEvidenceAsInt64(evidenceName string) (int64, error) {
	if a == nil {
		return 0, ErrEvidenceNotFound
	}

	var evidenceOut float64
	if err := a.GetEvidence(evidenceName, &evidenceOut); err != nil {
		return 0, err
	}

	for math.Mod(evidenceOut, 1) != 0 {
		evidenceOut *= 10
	}

	return int64(evidenceOut), nil
}

func (s Signals) GetSignal(signalName string, out interface{}) error {
	if s == nil {
		return ErrEvidenceNotFound
	}
	return getValueWithPath(jsonMap(s), signalName, out)
}

func (s Signals) GetSignalAsInt64(signalName string) (int64, error) {
	if s == nil {
		return 0, ErrEvidenceNotFound
	}

	var outFloat float64
	if err := s.GetSignal(signalName, &outFloat); err != nil {
		return 0, err
	}

	for math.Mod(outFloat, 1) != 0 {
		outFloat *= 10
	}

	return int64(outFloat), nil
}

// getValueWithPath navigates nested JSON objects using a dot-separated path ("a.b.c").
// It supports leaf values that are primitives or []interface{} (which can be bound to typed slices).
func getValueWithPath(root jsonMap, path string, out interface{}) error {
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
		return setToSlice(slice, out)
	}
	return setToPointer(v, out)
}

// keep error messages stable ("evidenceOut") to avoid breaking existing tests/clients.
func setToPointer(value interface{}, evidenceOut interface{}) error {
	evidenceOutReflectValue := reflect.ValueOf(evidenceOut)
	if evidenceOutReflectValue.Kind() != reflect.Ptr {
		return errors.New("expecting evidenceOut to be a pointer")
	}
	evidenceOutIndirectReflectKind := reflect.Indirect(evidenceOutReflectValue).Kind()

	valueReflectValue := reflect.ValueOf(value)
	valueReflectKind := valueReflectValue.Kind()

	if evidenceOutIndirectReflectKind != valueReflectKind {
		return fmt.Errorf("expecting evidenceOut to be a pointer to %s", valueReflectKind.String())
	}

	evidenceOutReflectValue.Elem().Set(valueReflectValue)
	return nil
}

// keep error messages stable ("evidenceOut") to avoid breaking existing tests/clients.
func setToSlice(evidenceSlice []interface{}, evidenceOut interface{}) error {
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
	Evidence       Evidence   `json:"evidence,omitempty"`
	Reasons        []Reason   `json:"reasons"`
	Signals        Signals    `json:"signals,omitempty"`
}

type TransactionAssessment struct {
	ID             string     `json:"id"`
	RiskAssessment Assessment `json:"risk_assessment"`
	DeviceID       string     `json:"device_id"`
	Evidence       Evidence   `json:"evidence,omitempty"`
	Reasons        []Reason   `json:"reasons"`
	Signals        Signals    `json:"signals,omitempty"`
}
