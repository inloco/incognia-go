package incognia

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LibMetricsTestSuite struct {
	suite.Suite
}

func (suite *LibMetricsTestSuite) TestEncodeDecodeWithAllValues() {
	original := &libMetrics{
		RequestID: "abc-123",
		Endpoint:  "/api/v2/onboarding/signups",
		Latency:   42,
	}

	encoded := original.encode()
	suite.NotEmpty(encoded)

	decoded, err := decodeLibMetrics(encoded)
	suite.NoError(err)
	suite.Equal(original.RequestID, decoded.RequestID)
	suite.Equal(original.Endpoint, decoded.Endpoint)
	suite.Equal(original.Latency, decoded.Latency)
}

func (suite *LibMetricsTestSuite) TestEncodeDecodeWithEmptyRequestID() {
	original := &libMetrics{
		RequestID: "",
		Endpoint:  "/api/v2/feedbacks",
		Latency:   17,
	}

	encoded := original.encode()
	suite.NotEmpty(encoded)

	decoded, err := decodeLibMetrics(encoded)
	suite.NoError(err)
	suite.Empty(decoded.RequestID)
	suite.Equal(original.Endpoint, decoded.Endpoint)
	suite.Equal(original.Latency, decoded.Latency)
}

func (suite *LibMetricsTestSuite) TestEncodeDecodeWithZeroLatency() {
	original := &libMetrics{
		RequestID: "req-999",
		Endpoint:  "/api/v2/authentication/transactions",
		Latency:   0,
	}

	encoded := original.encode()
	decoded, err := decodeLibMetrics(encoded)
	suite.NoError(err)
	suite.Equal(original.RequestID, decoded.RequestID)
	suite.Equal(original.Endpoint, decoded.Endpoint)
	suite.Equal(int64(0), decoded.Latency)
}

func (suite *LibMetricsTestSuite) TestEncodeDecodeWithAllEmpty() {
	original := &libMetrics{}

	encoded := original.encode()
	suite.NotEmpty(encoded)

	decoded, err := decodeLibMetrics(encoded)
	suite.NoError(err)
	suite.Empty(decoded.RequestID)
	suite.Empty(decoded.Endpoint)
	suite.Equal(int64(0), decoded.Latency)
}

func (suite *LibMetricsTestSuite) TestDecodeInvalidBase64ReturnsError() {
	_, err := decodeLibMetrics("not-valid-base64!!!")
	suite.Error(err)
}

func (suite *LibMetricsTestSuite) TestDecodeEmptyStringReturnsError() {
	_, err := decodeLibMetrics("")
	suite.Error(err)
}

func (suite *LibMetricsTestSuite) TestDecodeInvalidJSONReturnsError() {
	import64 := "aW52YWxpZC1qc29u" // base64 of "invalid-json"
	_, err := decodeLibMetrics(import64)
	suite.Error(err)
}

func TestLibMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(LibMetricsTestSuite))
}
