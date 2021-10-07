package incognia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessSignupAssessmentGetEvidence(t *testing.T) {
	expectedDeviceModel := "Moto Z2 Play"
	signupAssessment := &SignupAssessment{
		evidence: map[string]interface{}{
			"device_model": expectedDeviceModel,
		},
	}

	var deviceModel string

	err := signupAssessment.GetEvidence("device_model", &deviceModel)

	assert.NoError(t, err)
	assert.Equal(t, expectedDeviceModel, deviceModel)
}

func TestSuccessSignupAssessmentGetEvidenceAsInt64(t *testing.T) {
	expectedRiskWindowRemaining := float64(12.387238787)
	signupAssessment := &SignupAssessment{
		evidence: map[string]interface{}{
			"risk_window_remaining": expectedRiskWindowRemaining,
		},
	}

	riskWindowRemaining, err := signupAssessment.GetEvidenceAsInt64("risk_window_remaining")

	assert.NoError(t, err)
	assert.Equal(t, int64(expectedRiskWindowRemaining), riskWindowRemaining)
}
