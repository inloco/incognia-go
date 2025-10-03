package incognia

import "time"

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

type Location struct {
	Latitude    *float64   `json:"latitude"`
	Longitude   *float64   `json:"longitude"`
	CollectedAt *time.Time `json:"collected_at,omitempty"`
}

type postAssessmentRequestBody struct {
	InstallationID    string                 `json:"installation_id,omitempty"`
	RequestToken      string                 `json:"request_token,omitempty"`
	SessionToken      string                 `json:"session_token,omitempty"`
	AppVersion        string                 `json:"app_version,omitempty"`
	DeviceOs          string                 `json:"device_os,omitempty"`
	AddressLine       string                 `json:"address_line,omitempty"`
	StructuredAddress *StructuredAddress     `json:"structured_address,omitempty"`
	Coordinates       *Coordinates           `json:"address_coordinates,omitempty"`
	AccountID         string                 `json:"account_id,omitempty"`
	PolicyID          string                 `json:"policy_id,omitempty"`
	ExternalID        string                 `json:"external_id,omitempty"`
	CustomProperties  map[string]interface{} `json:"custom_properties,omitempty"`
	PersonID          *PersonID              `json:"person_id,omitempty"`
	DebtorAccount     *BankAccountInfo       `json:"debtor_account,omitempty"`
	CreditorAccount   *BankAccountInfo       `json:"creditor_account,omitempty"`
}

type FeedbackType string

const (
	AccountAllowed                    FeedbackType = "account_allowed"
	DeviceAllowed                     FeedbackType = "device_allowed"
	Verified                          FeedbackType = "verified"
	Reset                             FeedbackType = "reset"
	AccountTakeover                   FeedbackType = "account_takeover"
	IdentityFraud                     FeedbackType = "identity_fraud"
	Chargeback                        FeedbackType = "chargeback"
	ChargebackNotification            FeedbackType = "chargeback_notification"
	PromotionAbuse                    FeedbackType = "promotion_abuse"
	LoginAccepted                     FeedbackType = "login_accepted"
	LoginAcceptedByDeviceVerification FeedbackType = "login_accepted_by_device_verification"
	LoginAcceptedByFacialBiometrics   FeedbackType = "login_accepted_by_facial_biometrics"
	LoginAcceptedByManualReview       FeedbackType = "login_accepted_by_manual_review"
	LoginDeclined                     FeedbackType = "login_declined"
	LoginDeclinedByFacialBiometrics   FeedbackType = "login_declined_by_facial_biometrics"
	LoginDeclinedByManualReview       FeedbackType = "login_declined_by_manual_review"
	PaymentAccepted                   FeedbackType = "payment_accepted"
	PaymentAcceptedByControlGroup     FeedbackType = "payment_accepted_by_control_group"
	PaymentAcceptedByThirdParty       FeedbackType = "payment_accepted_by_third_party"
	PaymentDeclined                   FeedbackType = "payment_declined"
	PaymentDeclinedByAcquirer         FeedbackType = "payment_declined_by_acquirer"
	PaymentDeclinedByBusiness         FeedbackType = "payment_declined_by_business"
	PaymentDeclinedByManualReview     FeedbackType = "payment_declined_by_manual_review"
	PaymentDeclinedByRiskAnalysis     FeedbackType = "payment_declined_by_risk_analysis"
	SignupAccepted                    FeedbackType = "signup_accepted"
	SignupDeclined                    FeedbackType = "signup_declined"
)

type postFeedbackRequestBody struct {
	Event          FeedbackType `json:"event"`
	OccurredAt     *time.Time   `json:"occurred_at,omitempty"`
	ExpiresAt      *time.Time   `json:"expires_at,omitempty"`
	InstallationID string       `json:"installation_id,omitempty"`
	SessionToken   string       `json:"session_token,omitempty"`
	RequestToken   string       `json:"request_token,omitempty"`
	LoginID        string       `json:"login_id,omitempty"`
	PaymentID      string       `json:"payment_id,omitempty"`
	SignupID       string       `json:"signup_id,omitempty"`
	AccountID      string       `json:"account_id,omitempty"`
	ExternalID     string       `json:"external_id,omitempty"`
	PersonID       *PersonID    `json:"person_id,omitempty"`
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

type CouponType struct {
	Type        string  `json:"type"`
	Value       float64 `json:"value"`
	MaxDiscount float64 `json:"max_discount"`
	Id          string  `json:"id"`
	Name        string  `json:"name"`
}

type paymentMethodType string

const (
	AccountBalance paymentMethodType = "account_balance"
	ApplePay       paymentMethodType = "apple_pay"
	Bancolombia    paymentMethodType = "bancolombia"
	BoletoBancario paymentMethodType = "boleto_bancario"
	Cash           paymentMethodType = "cash"
	CreditCard     paymentMethodType = "credit_card"
	DebitCard      paymentMethodType = "debit_card"
	GooglePay      paymentMethodType = "google_pay"
	MealVoucher    paymentMethodType = "meal_voucher"
	NuPay          paymentMethodType = "nu_pay"
	Paypal         paymentMethodType = "paypal"
	Pix            paymentMethodType = "pix"
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
	Brand      string            `json:"brand,omitempty"`
}

type postTransactionRequestBody struct {
	ExternalID              string                 `json:"external_id,omitempty"`
	PolicyID                string                 `json:"policy_id,omitempty"`
	AppVersion              string                 `json:"app_version,omitempty"`
	Location                *Location              `json:"location,omitempty"`
	DeviceOs                string                 `json:"device_os,omitempty"`
	Coupon                  *CouponType            `json:"coupon,omitempty"`
	InstallationID          *string                `json:"installation_id,omitempty"`
	PaymentMethodIdentifier string                 `json:"payment_method_identifier,omitempty"`
	Type                    transactionType        `json:"type"`
	AccountID               string                 `json:"account_id"`
	Addresses               []*TransactionAddress  `json:"addresses,omitempty"`
	PaymentValue            *PaymentValue          `json:"payment_value,omitempty"`
	PaymentMethods          []*PaymentMethod       `json:"payment_methods,omitempty"`
	SessionToken            *string                `json:"session_token,omitempty"`
	RequestToken            string                 `json:"request_token,omitempty"`
	StoreID                 string                 `json:"store_id,omitempty"`
	CustomProperties        map[string]interface{} `json:"custom_properties,omitempty"`
	PersonID                *PersonID              `json:"person_id,omitempty"`
	DebtorAccount           *BankAccountInfo       `json:"debtor_account,omitempty"`
	CreditorAccount         *BankAccountInfo       `json:"creditor_account,omitempty"`
}

type PersonID struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type PixKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type BankAccountInfo struct {
	AccountType       string    `json:"account_type"`
	AccountPurpose    string    `json:"account_purpose"`
	HolderType        string    `json:"holder_type"`
	HolderTaxID       *PersonID `json:"holder_tax_id"`
	Country           string    `json:"country"`
	IspbCode          string    `json:"ispb_code"`
	BranchCode        string    `json:"branch_code"`
	AccountNumber     string    `json:"account_number"`
	AccountCheckDigit string    `json:"account_check_digit"`
	PixKeys           []*PixKey `json:"pix_keys"`
}
