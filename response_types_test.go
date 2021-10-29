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
