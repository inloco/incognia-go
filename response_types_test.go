package incognia

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

const evidencesJSON = `{"device_model": "Moto Z2 Play", "test_slice": ["something", "another_something"], "geocode_quality": "good", "last_location_ts": "2022-11-01T22:45:53.299Z", "address_quality": "good", "address_match": "street", "location_events_near_address": 38.0, "location_events_quantity": 0.0, "location_services": { "location_permission_enabled": true, "location_sensors_enabled": true }, "device_integrity": { "probable_root": false, "emulator": false, "gps_spoofing": false, "from_official_store": true }, "account_integrity": { "recent_high_risk_assessment": true, "risk_window_remaining": 199299292323}}`

type EvidenceTestSuite struct {
	suite.Suite

	Evidence Evidence
}

func (suite *EvidenceTestSuite) SetupTest() {
	var evidence Evidence

	err := json.Unmarshal([]byte(evidencesJSON), &evidence)
	suite.NoError(err)

	suite.Evidence = evidence
}

func (suite *EvidenceTestSuite) TestSuccessGetEvidenceAsInt64() {
	result, err := suite.Evidence.GetEvidenceAsInt64("account_integrity.risk_window_remaining")
	suite.NoError(err)
	suite.Equal(int64(199299292323), result)
}

func (suite *EvidenceTestSuite) TestSuccessGetEvidenceAsInt64Zero() {
	result, err := suite.Evidence.GetEvidenceAsInt64("location_events_quantity")
	suite.NoError(err)
	suite.Equal(int64(0), result)
}

func (suite *EvidenceTestSuite) TestSuccessGetEvidenceAsString() {
	var result string
	err := suite.Evidence.GetEvidence("device_model", &result)
	suite.NoError(err)
	suite.Equal("Moto Z2 Play", result)
}

func (suite *EvidenceTestSuite) TestSuccessGetEvidenceAsBool() {
	var result bool
	err := suite.Evidence.GetEvidence("location_services.location_permission_enabled", &result)
	suite.NoError(err)
	suite.True(result)
}

func (suite *EvidenceTestSuite) TestSuccessGetEvidenceAsFloat() {
	var result float64
	err := suite.Evidence.GetEvidence("location_events_near_address", &result)
	suite.NoError(err)
	suite.Equal(38.0, result)
}

func (suite *EvidenceTestSuite) TestSuccessGetEvidenceAsSlice() {
	var result []string
	err := suite.Evidence.GetEvidence("test_slice", &result)
	suite.NoError(err)
	suite.Equal([]string{"something", "another_something"}, result)
}

func (suite *EvidenceTestSuite) TestErrorGetEvidenceWrongType() {
	var result float64
	err := suite.Evidence.GetEvidence("geocode_quality", &result)
	suite.EqualError(err, "expecting evidenceOut to be a pointer to string")
}

func (suite *EvidenceTestSuite) TestErrorGetEvidenceWrongNumberType() {
	var result int64
	err := suite.Evidence.GetEvidence("location_events_near_address", &result)
	suite.EqualError(err, "expecting evidenceOut to be a pointer to float64")
}

func (suite *EvidenceTestSuite) TestErrorGetEvidenceNotFound() {
	var result int64
	err := suite.Evidence.GetEvidence("something", &result)
	suite.EqualError(err, ErrEvidenceNotFound.Error())
}

func (suite *EvidenceTestSuite) TestGetEvidence_WhenEvidenceNil_ReturnsNotFound() {
	var evidence Evidence

	var result string
	err := evidence.GetEvidence("device_model", &result)

	suite.EqualError(err, ErrEvidenceNotFound.Error())
}

func (suite *EvidenceTestSuite) TestUnmarshalAssessmentWithoutEvidenceDoesNotFail() {
	payload := []byte(`{
		"id":"1",
		"risk_assessment":"low_risk",
		"device_id":"device-1",
		"reasons":[]
	}`)

	var a TransactionAssessment
	err := json.Unmarshal(payload, &a)

	suite.NoError(err)
	suite.Nil(a.Evidence)
}

func (suite *EvidenceTestSuite) TestGetEvidenceAsInt64_WhenEvidenceNil_ReturnsNotFound() {
	var evidence Evidence

	_, err := evidence.GetEvidenceAsInt64("any.path")

	suite.EqualError(err, ErrEvidenceNotFound.Error())
}

func (suite *EvidenceTestSuite) TestErrorGetEvidenceNoPointer() {
	var result int64
	err := suite.Evidence.GetEvidence("location_events_near_address", result)
	suite.EqualError(err, "expecting evidenceOut to be a pointer")
}

func (suite *EvidenceTestSuite) TestErrorGetEvidenceWrongPath() {
	var result int64
	err := suite.Evidence.GetEvidence("location_events_near_address.something", result)
	suite.EqualError(err, ErrEvidenceNotFound.Error())
}

func TestEvidenceTestSuite(t *testing.T) {
	suite.Run(t, new(EvidenceTestSuite))
}

func (suite *EvidenceTestSuite) TestGetEvidenceAsInt64_WhenEvidenceNotFound_ReturnsError() {
	_, err := suite.Evidence.GetEvidenceAsInt64("app_tampering.does_not_exist")
	suite.Error(err, ErrEvidenceNotFound.Error())
}

func (suite *EvidenceTestSuite) TestGetEvidenceAsInt64_WhenEvidenceHasDecimal_MultipliesUntilInteger() {
	e := Evidence{
		"decimal_value": 1.5,
	}

	result, err := e.GetEvidenceAsInt64("decimal_value")
	suite.NoError(err)
	suite.Equal(int64(15), result)
}

func (suite *EvidenceTestSuite) TestGetEvidence_SliceOutNotPointer_ReturnsError() {
	e := Evidence{"arr": []interface{}{"a", "b"}}

	var out []string // <- slice, mas vamos passar sem ponteiro
	err := e.GetEvidence("arr", out)

	suite.EqualError(err, "expecting evidenceOut to be a pointer to slice")
}

func (suite *EvidenceTestSuite) TestGetEvidence_SliceOutPointerButNotSlice_ReturnsError() {
	e := Evidence{"arr": []interface{}{"a", "b"}}

	var out int
	err := e.GetEvidence("arr", &out)

	suite.EqualError(err, "expecting evidenceOut to be a pointer to slice")
}

const signalsJSON = `{
  "installation": {
    "first_assessment_request": {
      "duration_since": "PT3954H45M20.377085441S",
      "timestamp": "2025-06-26T21:35:10.547Z"
    },
    "app_debugging": "detected",
    "has_device_id": true
  },
  "device": {
    "emulator": "detected",
    "root": "detected",
	"accessed_accounts_3d": 4
  }
}`

type SignalsTestSuite struct {
	suite.Suite
	Signals Signals
}

func (suite *SignalsTestSuite) SetupTest() {
	var s Signals
	err := json.Unmarshal([]byte(signalsJSON), &s)
	suite.NoError(err)
	suite.Signals = s
}

func (suite *SignalsTestSuite) TestSuccessGetSignalAsString() {
	var result string
	err := suite.Signals.GetSignal("device.emulator", &result)
	suite.NoError(err)
	suite.Equal("detected", result)
}

func (suite *SignalsTestSuite) TestSuccessGetSignalAsBool() {
	var result bool
	err := suite.Signals.GetSignal("installation.has_device_id", &result)
	suite.NoError(err)
	suite.True(result)
}

func (suite *SignalsTestSuite) TestSuccessGetSignalNestedTimestamp() {
	var result string
	err := suite.Signals.GetSignal("installation.first_assessment_request.timestamp", &result)
	suite.NoError(err)
	suite.Equal("2025-06-26T21:35:10.547Z", result)
}

func (suite *SignalsTestSuite) TestGetSignal_WhenSignalsNil_ReturnsNotFound() {
	var s Signals
	var result string
	err := s.GetSignal("device.emulator", &result)
	suite.Error(err, ErrEvidenceNotFound.Error())
}

func TestSignalsTestSuite(t *testing.T) {
	suite.Run(t, new(SignalsTestSuite))
}

func (suite *SignalsTestSuite) TestSuccessGetSignalAsInt64() {
	result, err := suite.Signals.GetSignalAsInt64("device.accessed_accounts_3d")
	suite.NoError(err)
	suite.Equal(int64(4), result)
}

func (suite *SignalsTestSuite) TestGetSignalAsInt64_WhenSignalsNil_ReturnsNotFound() {
	var s Signals
	_, err := s.GetSignalAsInt64("device.accessed_accounts_3d")
	suite.Error(err, ErrEvidenceNotFound.Error())
}

func (suite *SignalsTestSuite) TestGetSignalAsInt64_WhenSignalNotFound_ReturnsError() {
	_, err := suite.Signals.GetSignalAsInt64("device.does_not_exist")
	suite.Error(err, ErrEvidenceNotFound.Error())
}
func (suite *SignalsTestSuite) TestGetSignalAsInt64_WhenSignalHasDecimal_MultipliesUntilInteger() {
	s := Signals{
		"decimal_value": 1.5,
	}

	result, err := s.GetSignalAsInt64("decimal_value")
	suite.NoError(err)
	suite.Equal(int64(15), result)
}
