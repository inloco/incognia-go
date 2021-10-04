package incognia

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type StructuredAddress struct {
	Locale       string `json:"locale"`
	CountryName  string `json:"country_name"`
	CountryCode  string `json:"country_code"`
	State        string `json:"state"`
	City         string `json:"city"`
	Borough      string `json:"borough"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complements  string `json:"complements"`
	PostalCode   string `json:"postal_code"`
}

type Address struct {
	Coordinates       *Coordinates
	StructuredAddress *StructuredAddress
	AddressLine       string
}

type postAssessmentRequestBody struct {
	InstallationID    string             `json:"installation_id"`
	AddressLine       string             `json:"address_line,omitempty"`
	StructuredAddress *StructuredAddress `json:"structured_address,omitempty"`
	Coordinates       *Coordinates       `json:"address_coordinates,omitempty"`
}

type FeedbackType string

const (
	PaymentAccepted               FeedbackType = "payment_accepted"
	PaymentDeclined               FeedbackType = "payment_declined"
	PaymentDeclinedByRiskAnalysis FeedbackType = "payment_declined_by_risk_analysis"
	PaymentDeclinedByAcquirer     FeedbackType = "payment_declined_by_acquirer"
	PaymentDeclinedByBusiness     FeedbackType = "payment_declined_by_business"
	PaymentDeclinedByManualReview FeedbackType = "payment_declined_by_manual_review"
	LoginAccepted                 FeedbackType = "login_accepted"
	LoginDeclined                 FeedbackType = "login_declined"
	SignupAccepted                FeedbackType = "signup_accepted"
	SignupDeclined                FeedbackType = "signup_declined"
	ChallengePassed               FeedbackType = "challenge_passed"
	ChallengeFailed               FeedbackType = "challenge_failed"
	PasswordChangedSuccessfully   FeedbackType = "password_changed_successfully"
	PasswordChangeFailed          FeedbackType = "password_change_failed"
	Verified                      FeedbackType = "verified"
	NotVerified                   FeedbackType = "not_verified"
	Chargeback                    FeedbackType = "chargeback"
	PromotionAbuse                FeedbackType = "promotion_abuse"
	AccountTakeover               FeedbackType = "account_takeover"
	MposFraud                     FeedbackType = "mpos_fraud"
)

type FeedbackIdentifiers struct {
	InstallationID string
	LoginID        string
	PaymentID      string
	SignupID       string
	AccountID      string
	ExternalID     string
}

type postFeedbackRequestBody struct {
	Event          FeedbackType `json:"event"`
	Timestamp      int64        `json:"timestamp"`
	InstallationID string       `json:"installation_id,omitempty"`
	LoginID        string       `json:"login_id,omitempty"`
	PaymentID      string       `json:"payment_id,omitempty"`
	SignupID       string       `json:"signup_id,omitempty"`
	AccountID      string       `json:"account_id,omitempty"`
	ExternalID     string       `json:"external_id,omitempty"`
}
