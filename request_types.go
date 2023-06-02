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
	PaymentAcceptedByThirdParty   FeedbackType = "payment_accepted_by_third_party"
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
	ChargebackNotification        FeedbackType = "chargeback_notification"
)

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

type AddressType string

const (
	Shipping AddressType = "shipping"
	Billing  AddressType = "billing"
	Home     AddressType = "home"
)

type transactionType string

const (
	loginType   transactionType = "login"
	paymentType transactionType = "payment"
)

type TransactionAddress struct {
	Type              AddressType        `json:"type"`
	Coordinates       *Coordinates       `json:"address_coordinates"`
	StructuredAddress *StructuredAddress `json:"structured_address"`
	AddressLine       string             `json:"address_line"`
}

type PaymentValue struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type paymentMethodType string

const (
	CreditCard     paymentMethodType = "credit_card"
	DebitCard      paymentMethodType = "debit_card"
	GooglePay      paymentMethodType = "google_pay"
	ApplePay       paymentMethodType = "apple_pay"
	NuPay          paymentMethodType = "nu_pay"
	Pix            paymentMethodType = "pix"
	MealVoucher    paymentMethodType = "meal_voucher"
	AccountBalance paymentMethodType = "account_balance"
)

type CardInfo struct {
	Bin            string `json:"bin"`
	LastFourDigits string `json:"last_four_digits"`
	ExpiryYear     string `json:"expiry_year,omitempty"`
	ExpiryMonth    string `json:"expiry_month,omitempty"`
}

type PaymentMethod struct {
	Identifier string            `json:"identifier,omitempty"`
	Type       paymentMethodType `json:"type"`
	CreditCard *CardInfo         `json:"credit_card_info,omitempty"`
	DebitCard  *CardInfo         `json:"debit_card_info,omitempty"`
}

type postTransactionRequestBody struct {
	ExternalID              string                `json:"external_id,omitempty"`
	PolicyID                string                `json:"policy_id,omitempty"`
	InstallationID          string                `json:"installation_id"`
	PaymentMethodIdentifier string                `json:"payment_method_identifier,omitempty"`
	Type                    transactionType       `json:"type"`
	AccountID               string                `json:"account_id"`
	Addresses               []*TransactionAddress `json:"addresses,omitempty"`
	PaymentValue            *PaymentValue         `json:"payment_value,omitempty"`
	PaymentMethods          []*PaymentMethod      `json:"payment_methods,omitempty"`
}
