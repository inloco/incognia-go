package incognia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	clientID       string = "client-id"
	clientSecret   string = "client-secret"
	token          string = "some-token"
	tokenExpiresIn string = "500"
)

var (
	userAgentRegex   = regexp.MustCompile(`^incognia-api-go(/(v[0-9]+\.[0-9]+\.[0-9]+|unknown))? \([a-z]+ [a-z0-9]+\) Go/go[0-9]+\.[0-9]+\.[0-9]+$`)
	now              = time.Now()
	floatVar         = -7.5432
	FixedCollectedAt = time.Date(2025, time.March, 22, 12, 12, 12, 0, time.UTC)
	nowMinusSeconds  = now.Add(-1 * time.Second)
	installationId   = "installation-id"
	requestToken     = "request-token"
	customProperty   = map[string]interface{}{
		"custom_1": "custom_value_1",
		"custom_2": "custom_value_2",
	}
	shouldEval               bool                = true
	shouldNotEval            bool                = false
	emptyQueryString         map[string][]string = nil
	queryStringWithFalseEval                     = map[string][]string{"eval": []string{"false"}}
	queryStringWithTrueEval                      = map[string][]string{"eval": []string{"true"}}
	customPropertiesFixture                      = map[string]interface{}{
		"user_id":       "a9f7e3b2-24cd-4d3e-b6e1-98d1234c5678",
		"is_verified":   true,
		"last_latitude": -23.55052,
		"preferences": map[string]interface{}{
			"notifications_enabled": true,
			"language":              "en-US",
		},
		"metadata": nil,
	}
	pixKeyArrayFixture = []*PixKey{
		{Type: "cpf", Value: "12345678901"},
		{Type: "email", Value: "legit_person@gmail.com"},
	}
	bankAccountInfoFixture = &BankAccountInfo{
		AccountType:       "savings",
		AccountPurpose:    "rural",
		HolderType:        "business",
		HolderTaxID:       &PersonID{Type: "cpf", Value: "12345678901"},
		Country:           "BR",
		IspbCode:          "18236120",
		BranchCode:        "0001",
		AccountNumber:     "123456",
		AccountCheckDigit: "0",
		PixKeys:           pixKeyArrayFixture,
	}
	locationFixtureFull = &Location{
		Latitude:    &floatVar,
		Longitude:   &floatVar,
		CollectedAt: &FixedCollectedAt,
	}
	locationFixtureMissingLat = &Location{
		Latitude:    nil,
		Longitude:   &floatVar,
		CollectedAt: &FixedCollectedAt,
	}
	locationFixtureMissingLong = &Location{
		Latitude:    &floatVar,
		Longitude:   nil,
		CollectedAt: &FixedCollectedAt,
	}
	locationFixtureMissingCollectedAt = &Location{
		Latitude:    &floatVar,
		Longitude:   &floatVar,
		CollectedAt: nil,
	}
	signupAssessmentFixture = &SignupAssessment{
		ID:             "some-id",
		DeviceID:       "some-device-id",
		RequestID:      "some-request-id",
		RiskAssessment: LowRisk,
		Reasons:        []Reason{{Code: "mpos_fraud", Source: "global"}, {Code: "mpos_fraud", Source: "local"}},
		Evidence: Evidence{
			"device_model":                 "Moto Z2 Play",
			"geocode_quality":              "good",
			"address_quality":              "good",
			"address_match":                "street",
			"location_events_near_address": 38.0,
			"location_events_quantity":     288.0,
			"location_services": map[string]interface{}{
				"location_permission_enabled": true,
				"location_sensors_enabled":    true,
			},
			"device_integrity": map[string]interface{}{
				"probable_root":       false,
				"emulator":            false,
				"gps_spoofing":        false,
				"from_official_store": true,
			},
		},
	}
	postSignupRequestBodyFixture = &postAssessmentRequestBody{
		InstallationID: installationId,
		AddressLine:    "address line",
		StructuredAddress: &StructuredAddress{
			Locale:       "locale",
			CountryName:  "country-name",
			CountryCode:  "country-code",
			State:        "state",
			City:         "city",
			Borough:      "borough",
			Neighborhood: "neighborhood",
			Street:       "street",
			Number:       "number",
			Complements:  "complements",
			PostalCode:   "postalcode",
		},
		Coordinates: &Coordinates{
			Lat: -23.561414,
			Lng: -46.6558819,
		},
	}

	postSignupRequestBodyWithAllParamsFixture = &postAssessmentRequestBody{
		InstallationID:   installationId,
		RequestToken:     requestToken,
		AddressLine:      "address line",
		DeviceOs:         "ios",
		AppVersion:       "1.2.3",
		CustomProperties: customPropertiesFixture,
		StructuredAddress: &StructuredAddress{
			Locale:       "locale",
			CountryName:  "country-name",
			CountryCode:  "country-code",
			State:        "state",
			City:         "city",
			Borough:      "borough",
			Neighborhood: "neighborhood",
			Street:       "street",
			Number:       "number",
			Complements:  "complements",
			PostalCode:   "postalcode",
		},
		Coordinates: &Coordinates{
			Lat: -23.561414,
			Lng: -46.6558819,
		},
		AccountID:  "account-id",
		PolicyID:   "policy-id",
		ExternalID: "external-id",
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	postWebSignupRequestBodyWithAllParamsFixture = &postAssessmentRequestBody{
		RequestToken:     requestToken,
		AccountID:        "account-id",
		PolicyID:         "policy-id",
		CustomProperties: customPropertiesFixture,
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	postWebSignupRequestBodyFixture = &postAssessmentRequestBody{
		RequestToken: requestToken,
		PolicyID:     "policy-id",
	}
	postWebSignupRequestBodyRequiredFieldsFixture = &postAssessmentRequestBody{
		RequestToken: requestToken,
	}
	postWebSignupRequestBodyMissingRequestTokenFixture = &postAssessmentRequestBody{
		PolicyID: "policy-id",
	}
	postSignupRequestBodyRequiredFieldsFixture = &postAssessmentRequestBody{
		InstallationID: installationId,
	}
	addressFixture = &Address{
		Coordinates:       postSignupRequestBodyFixture.Coordinates,
		StructuredAddress: postSignupRequestBodyFixture.StructuredAddress,
		AddressLine:       postSignupRequestBodyFixture.AddressLine,
	}
	postFeedbackRequestBodyFixture = &postFeedbackRequestBody{
		Event:          SignupAccepted,
		OccurredAt:     &now,
		InstallationID: "some-installation-id",
		SessionToken:   "some-session-token",
		RequestToken:   "some-request-token",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	postFeedbackRequestWithExpirationBodyFixture = &postFeedbackRequestBody{
		Event:          SignupAccepted,
		OccurredAt:     &now,
		ExpiresAt:      &nowMinusSeconds,
		InstallationID: "some-installation-id",
		SessionToken:   "some-session-token",
		RequestToken:   "some-request-token",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
	}
	postFeedbackRequestBodyRequiredFieldsFixture = &postFeedbackRequestBody{
		Event: SignupAccepted,
	}
	feedbackIdentifiersFixture = &FeedbackIdentifiers{
		InstallationID: "some-installation-id",
		SessionToken:   "some-session-token",
		RequestToken:   "some-request-token",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	emptyTransactionAssessmentFixture = &TransactionAssessment{}
	transactionAssessmentFixture      = &TransactionAssessment{
		ID:             "some-id",
		DeviceID:       "some-device-id",
		RiskAssessment: LowRisk,
		Reasons:        []Reason{{Code: "mpos_fraud", Source: "global"}, {Code: "mpos_fraud", Source: "local"}},
		Evidence: Evidence{
			"device_model":                 "Moto Z2 Play",
			"geocode_quality":              "good",
			"address_quality":              "good",
			"address_match":                "street",
			"location_events_near_address": 38.0,
			"location_events_quantity":     288.0,
			"location_services": map[string]interface{}{
				"location_permission_enabled": true,
				"location_sensors_enabled":    true,
			},
			"device_integrity": map[string]interface{}{
				"probable_root":       false,
				"emulator":            false,
				"gps_spoofing":        false,
				"from_official_store": true,
			},
		},
	}
	postPaymentRequestBodyFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		DeviceOs:       "android",
		AppVersion:     "1.2.3",
		PolicyID:       "policy-id",
		Type:           paymentType,
		Coupon: &CouponType{
			Type:        "coupon_type",
			Value:       55.02,
			MaxDiscount: 30,
			Id:          "identifier",
			Name:        "CouponName",
		},
		StoreID:          "store-id",
		CustomProperties: customProperty,
		Addresses: []*TransactionAddress{
			{
				Type: Billing,
				StructuredAddress: &StructuredAddress{
					Locale:       "locale",
					CountryName:  "country-name",
					CountryCode:  "country-code",
					State:        "state",
					City:         "city",
					Borough:      "borough",
					Neighborhood: "neighborhood",
					Street:       "street",
					Number:       "number",
					Complements:  "complements",
					PostalCode:   "postalcode",
				},
				AddressLine: "address line",
				Coordinates: &Coordinates{
					Lat: -23.561414,
					Lng: -46.6558819,
				},
			},
		},
		PaymentValue: &PaymentValue{
			Amount:   55.02,
			Currency: "BRL",
		},
		PaymentMethods: []*PaymentMethod{
			{
				Type:       CreditCard,
				Identifier: "credit-card-hash-123",
				Brand:      "visa",
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
		DebtorAccount:   bankAccountInfoFixture,
		CreditorAccount: bankAccountInfoFixture,
	}
	postPaymentWebRequestBodyFixture = &postTransactionRequestBody{
		RequestToken: requestToken,
		AccountID:    "account-id",
		ExternalID:   "external-id",
		PolicyID:     "policy-id",
		Coupon: &CouponType{
			Type:        "coupon_type",
			Value:       55.02,
			MaxDiscount: 30,
			Id:          "identifier",
			Name:        "CouponName",
		},
		Type: paymentType,
		Addresses: []*TransactionAddress{
			{
				Type: Billing,
				StructuredAddress: &StructuredAddress{
					Locale:       "locale",
					CountryName:  "country-name",
					CountryCode:  "country-code",
					State:        "state",
					City:         "city",
					Borough:      "borough",
					Neighborhood: "neighborhood",
					Street:       "street",
					Number:       "number",
					Complements:  "complements",
					PostalCode:   "postalcode",
				},
				AddressLine: "address line",
				Coordinates: &Coordinates{
					Lat: -23.561414,
					Lng: -46.6558819,
				},
			},
		},
		PaymentValue: &PaymentValue{
			Amount:   55.02,
			Currency: "BRL",
		},
		PaymentMethods: []*PaymentMethod{
			{
				Type:       CreditCard,
				Identifier: "credit-card-hash-123",
				Brand:      "visa",
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	postPaymentRequestBodyRequiredFieldsFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		Type:           paymentType,
	}
	paymentFixture = &Payment{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		StoreID:        "store-id",
		DeviceOs:       "android",
		AppVersion:     "1.2.3",
		Coupon: &CouponType{
			Type:        "coupon_type",
			Value:       55.02,
			MaxDiscount: 30,
			Id:          "identifier",
			Name:        "CouponName",
		},
		CustomProperties: customProperty,
		Addresses: []*TransactionAddress{
			{
				Type: Billing,
				StructuredAddress: &StructuredAddress{
					Locale:       "locale",
					CountryName:  "country-name",
					CountryCode:  "country-code",
					State:        "state",
					City:         "city",
					Borough:      "borough",
					Neighborhood: "neighborhood",
					Street:       "street",
					Number:       "number",
					Complements:  "complements",
					PostalCode:   "postalcode",
				},
				AddressLine: "address line",
				Coordinates: &Coordinates{
					Lat: -23.561414,
					Lng: -46.6558819,
				},
			},
		},
		Value: &PaymentValue{
			Amount:   55.02,
			Currency: "BRL",
		},
		Methods: []*PaymentMethod{
			{
				Type:       CreditCard,
				Identifier: "credit-card-hash-123",
				Brand:      "visa",
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
		DebtorAccount:   bankAccountInfoFixture,
		CreditorAccount: bankAccountInfoFixture,
	}
	paymentWebFixture = &Payment{
		RequestToken: requestToken,
		AccountID:    "account-id",
		ExternalID:   "external-id",
		PolicyID:     "policy-id",
		Coupon: &CouponType{
			Type:        "coupon_type",
			Value:       55.02,
			MaxDiscount: 30,
			Id:          "identifier",
			Name:        "CouponName",
		},
		Addresses: []*TransactionAddress{
			{
				Type: Billing,
				StructuredAddress: &StructuredAddress{
					Locale:       "locale",
					CountryName:  "country-name",
					CountryCode:  "country-code",
					State:        "state",
					City:         "city",
					Borough:      "borough",
					Neighborhood: "neighborhood",
					Street:       "street",
					Number:       "number",
					Complements:  "complements",
					PostalCode:   "postalcode",
				},
				AddressLine: "address line",
				Coordinates: &Coordinates{
					Lat: -23.561414,
					Lng: -46.6558819,
				},
			},
		},
		Value: &PaymentValue{
			Amount:   55.02,
			Currency: "BRL",
		},
		Methods: []*PaymentMethod{
			{
				Type:       CreditCard,
				Identifier: "credit-card-hash-123",
				Brand:      "visa",
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	paymentFixtureRequiredFields = &Payment{
		InstallationID: &installationId,
		AccountID:      "account-id",
	}
	paymentFixtureWithLocation = &Payment{
		InstallationID: &installationId,
		AccountID:      "account-id",
		Location:       nil,
	}
	simplePaymentFixtureWithShouldEval = &Payment{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Eval:           &shouldEval,
	}
	simplePaymentFixtureWithShouldNotEval = &Payment{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Eval:           &shouldNotEval,
	}
	postSimplePaymentRequestBodyFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Type:           paymentType,
	}
	loginFixture = &Login{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		DeviceOs:                "Android",
		AppVersion:              "1.2.3",
		CustomProperties:        customProperty,
		PaymentMethodIdentifier: "payment-method-identifier",
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	postPaymentRequestBodyWithLocationFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		Type:           paymentType,
		Location:       nil,
	}
	loginFixtureWithShouldEval = &Login{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		DeviceOs:                "Android",
		AppVersion:              "1.2.3",
		PaymentMethodIdentifier: "payment-method-identifier",
		Eval:                    &shouldEval,
		CustomProperties:        customProperty,
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	loginFixtureWithShouldNotEval = &Login{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Eval:           &shouldNotEval,
	}
	loginWebFixture = &WebLogin{
		AccountID:        "account-id",
		ExternalID:       "external-id",
		PolicyID:         "policy-id",
		RequestToken:     requestToken,
		CustomProperties: customProperty,
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	loginWebFixtureWithShouldEval = &WebLogin{
		AccountID:        "account-id",
		ExternalID:       "external-id",
		PolicyID:         "policy-id",
		RequestToken:     requestToken,
		Eval:             &shouldEval,
		CustomProperties: customProperty,
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	loginWebFixtureWithShouldNotEval = &WebLogin{
		AccountID:        "account-id",
		ExternalID:       "external-id",
		PolicyID:         "policy-id",
		RequestToken:     requestToken,
		Eval:             &shouldNotEval,
		CustomProperties: customProperty,
	}
	loginFixtureWithLocation = &Login{
		InstallationID: &installationId,
		AccountID:      "account-id",
		Location:       nil,
	}
	postLoginRequestBodyFixture = &postTransactionRequestBody{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		DeviceOs:                "android",
		AppVersion:              "1.2.3",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		Type:                    loginType,
		CustomProperties:        customProperty,
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	postLoginWebRequestBodyFixture = &postTransactionRequestBody{
		AccountID:        "account-id",
		ExternalID:       "external-id",
		PolicyID:         "policy-id",
		Type:             loginType,
		RequestToken:     requestToken,
		CustomProperties: customProperty,
		PersonID: &PersonID{
			Type:  "cpf",
			Value: "12345678901",
		},
	}
	postLoginRequestBodyWithLocationFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		Type:           loginType,
		Location:       nil,
	}
)

type PanickingTokenProvider struct {
	panicString string
}

func (tokenProvider PanickingTokenProvider) GetToken() (Token, error) {
	panic(tokenProvider.panicString)
}

type IncogniaTestSuite struct {
	suite.Suite

	client      *Client
	token       string
	tokenServer *httptest.Server
}

func (suite *IncogniaTestSuite) SetupTest() {
	client, _ := New(&IncogniaClientConfig{ClientID: clientID, ClientSecret: clientSecret})
	suite.client = client

	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	suite.client.tokenProvider.(*AutoRefreshTokenProvider).tokenClient.tokenEndpoint = tokenServer.URL

	suite.client.endpoints.Token = tokenServer.URL
	suite.token = token
	suite.tokenServer = tokenServer
}

func (suite *IncogniaTestSuite) TearDownTest() {
	defer suite.tokenServer.Close()
}

func (suite *IncogniaTestSuite) TestManualRefreshTokenProviderErrorTokenNotFound() {
	tokenProvider := NewManualRefreshTokenProvider(NewTokenClient(&TokenClientConfig{ClientID: clientID, ClientSecret: clientSecret}))
	client, _ := New(&IncogniaClientConfig{ClientID: clientID, ClientSecret: clientSecret, TokenProvider: tokenProvider})

	_, err := client.RegisterLogin(loginFixture)
	suite.EqualError(err, ErrTokenNotFound.Error())

	_, err = client.RegisterPayment(paymentFixture)
	suite.EqualError(err, ErrTokenNotFound.Error())

	err = client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.EqualError(err, ErrTokenNotFound.Error())
}

func (suite *IncogniaTestSuite) TestManualRefreshTokenProviderSuccess() {
	tokenProvider := NewManualRefreshTokenProvider(NewTokenClient(&TokenClientConfig{ClientID: clientID, ClientSecret: clientSecret}))
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	tokenProvider.tokenClient.tokenEndpoint = tokenServer.URL
	client, _ := New(&IncogniaClientConfig{ClientID: clientID, ClientSecret: clientSecret, TokenProvider: tokenProvider})

	tokenProvider.Refresh()

	suite.client = client

	loginServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer loginServer.Close()
	_, err := client.RegisterLogin(loginFixture)
	suite.NoError(err)

	paymentServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer paymentServer.Close()
	_, err = client.RegisterPayment(paymentFixture)
	suite.NoError(err)

	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()
	err = client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupWithParams() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyWithAllParamsFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignupWithParams(&Signup{
		InstallationID:   postSignupRequestBodyWithAllParamsFixture.InstallationID,
		RequestToken:     postSignupRequestBodyWithAllParamsFixture.RequestToken,
		SessionToken:     postSignupRequestBodyWithAllParamsFixture.SessionToken,
		DeviceOs:         postSignupRequestBodyWithAllParamsFixture.DeviceOs,
		AppVersion:       postSignupRequestBodyWithAllParamsFixture.AppVersion,
		Address:          addressFixture,
		AccountID:        postSignupRequestBodyWithAllParamsFixture.AccountID,
		PolicyID:         postSignupRequestBodyWithAllParamsFixture.PolicyID,
		ExternalID:       postSignupRequestBodyWithAllParamsFixture.ExternalID,
		CustomProperties: postSignupRequestBodyWithAllParamsFixture.CustomProperties,
		PersonID:         postSignupRequestBodyWithAllParamsFixture.PersonID,
	})
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterWebSignupFull() {
	signupServer := suite.mockPostSignupsEndpoint(token, postWebSignupRequestBodyWithAllParamsFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterWebSignup(&WebSignup{
		RequestToken:     postWebSignupRequestBodyWithAllParamsFixture.RequestToken,
		AccountID:        postWebSignupRequestBodyWithAllParamsFixture.AccountID,
		PolicyID:         postWebSignupRequestBodyWithAllParamsFixture.PolicyID,
		CustomProperties: postWebSignupRequestBodyWithAllParamsFixture.CustomProperties,
		PersonID:         postWebSignupRequestBodyWithAllParamsFixture.PersonID,
	})
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignup() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupNilOptional() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyRequiredFieldsFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyRequiredFieldsFixture.InstallationID, nil)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterWebSignupNilOptional() {
	signupServer := suite.mockPostSignupsEndpoint(token, postWebSignupRequestBodyRequiredFieldsFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterWebSignup(&WebSignup{
		RequestToken: postWebSignupRequestBodyRequiredFieldsFixture.RequestToken,
	})
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupAfterTokenExpiration() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterSignupEmptyInstallationId() {
	response, err := suite.client.RegisterSignup("", &Address{})
	suite.EqualError(err, ErrMissingIdentifier.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterWebSignupEmptyRequestToken() {
	response, err := suite.client.RegisterWebSignup(&WebSignup{
		PolicyID: postWebSignupRequestBodyMissingRequestTokenFixture.PolicyID,
	})
	suite.EqualError(err, ErrMissingIdentifier.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterSignup() {
	signupServer := suite.mockPostSignupsEndpoint("some-other-token", postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestRegisterSignupErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Signups = statusServer.URL

		response, err := suite.client.RegisterSignup("any-signup-id", &Address{})
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedback() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackNilOptional() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyRequiredFieldsFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedback(postFeedbackRequestBodyRequiredFieldsFixture.Event, nil, nil)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackAfterTokenExpiration() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.NoError(err)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	err = suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterFeedback() {
	feedbackServer := suite.mockFeedbackEndpoint("some-other-token", postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestErrorsRegisterFeedback() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Feedback = statusServer.URL

		err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackWithExpiration() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestWithExpirationBodyFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedbackWithExpiration(postFeedbackRequestWithExpirationBodyFixture.Event, postFeedbackRequestWithExpirationBodyFixture.OccurredAt, postFeedbackRequestWithExpirationBodyFixture.ExpiresAt, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackWithExpirationNilOptional() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyRequiredFieldsFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedbackWithExpiration(postFeedbackRequestBodyRequiredFieldsFixture.Event, nil, nil, nil)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackWithExpirationAfterTokenExpiration() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestWithExpirationBodyFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedbackWithExpiration(postFeedbackRequestWithExpirationBodyFixture.Event, postFeedbackRequestWithExpirationBodyFixture.OccurredAt, postFeedbackRequestWithExpirationBodyFixture.ExpiresAt, feedbackIdentifiersFixture)
	suite.NoError(err)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	err = suite.client.RegisterFeedbackWithExpiration(postFeedbackRequestWithExpirationBodyFixture.Event, postFeedbackRequestWithExpirationBodyFixture.OccurredAt, postFeedbackRequestWithExpirationBodyFixture.ExpiresAt, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterFeedbackWithExpiration() {
	feedbackServer := suite.mockFeedbackEndpoint("some-other-token", postFeedbackRequestWithExpirationBodyFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedbackWithExpiration(postFeedbackRequestWithExpirationBodyFixture.Event, postFeedbackRequestWithExpirationBodyFixture.OccurredAt, postFeedbackRequestWithExpirationBodyFixture.ExpiresAt, feedbackIdentifiersFixture)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestErrorsRegisterFeedbackWithExpiration() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Feedback = statusServer.URL

		err := suite.client.RegisterFeedbackWithExpiration(postFeedbackRequestWithExpirationBodyFixture.Event, postFeedbackRequestWithExpirationBodyFixture.OccurredAt, postFeedbackRequestWithExpirationBodyFixture.ExpiresAt, feedbackIdentifiersFixture)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPayment() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentWeb() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentWebRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentWebFixture)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentNilOptional() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyRequiredFieldsFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixtureRequiredFields)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentAfterTokenExpiration() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.RegisterPayment(paymentFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentNilPayment() {
	response, err := suite.client.RegisterPayment(nil)
	suite.EqualError(err, ErrMissingPayment.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentEmptyInstallationId() {
	response, err := suite.client.RegisterPayment(&Payment{AccountID: "some-account-id"})
	suite.EqualError(err, ErrMissingIdentifier.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentEmptyAccountId() {
	response, err := suite.client.RegisterPayment(&Payment{InstallationID: &installationId})
	suite.EqualError(err, ErrMissingAccountID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterPayment() {
	transactionServer := suite.mockPostTransactionsEndpoint("some-other-token", postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestRegisterPaymentErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Transactions = statusServer.URL

		response, err := suite.client.RegisterPayment(paymentFixture)
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentWithEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postSimplePaymentRequestBodyFixture, transactionAssessmentFixture, queryStringWithTrueEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(simplePaymentFixtureWithShouldEval)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentWithFalseEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postSimplePaymentRequestBodyFixture, transactionAssessmentFixture, queryStringWithFalseEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(simplePaymentFixtureWithShouldNotEval)
	suite.NoError(err)
	suite.Equal(emptyTransactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentWithLocationAndTimestamp() {
	paymentFixtureWithLocation.Location = locationFixtureFull
	postPaymentRequestBodyWithLocationFixture.Location = locationFixtureFull

	*postPaymentRequestBodyWithLocationFixture.Location.CollectedAt = postPaymentRequestBodyWithLocationFixture.Location.CollectedAt.Round(0)
	*paymentFixtureWithLocation.Location.CollectedAt = paymentFixtureWithLocation.Location.CollectedAt.Round(0)

	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyWithLocationFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixtureWithLocation)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentWithLocationWithoutTimestamp() {
	paymentFixtureWithLocation.Location = locationFixtureMissingCollectedAt
	postPaymentRequestBodyWithLocationFixture.Location = locationFixtureMissingCollectedAt

	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyWithLocationFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixtureWithLocation)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentWithLocationMissingLat() {
	paymentFixtureWithLocation.Location = locationFixtureMissingLat
	postPaymentRequestBodyWithLocationFixture.Location = locationFixtureMissingLat

	*postPaymentRequestBodyWithLocationFixture.Location.CollectedAt = postPaymentRequestBodyWithLocationFixture.Location.CollectedAt.Round(0)
	*paymentFixtureWithLocation.Location.CollectedAt = paymentFixtureWithLocation.Location.CollectedAt.Round(0)

	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyWithLocationFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixtureWithLocation)

	suite.Nil(response)
	suite.EqualError(err, ErrMissingLocationLatLong.Error())
}

func (suite *IncogniaTestSuite) TestRegisterPaymentWithLocationMissingLong() {
	paymentFixtureWithLocation.Location = locationFixtureMissingLong
	postPaymentRequestBodyWithLocationFixture.Location = locationFixtureMissingLong

	*postPaymentRequestBodyWithLocationFixture.Location.CollectedAt = postPaymentRequestBodyWithLocationFixture.Location.CollectedAt.Round(0)
	*paymentFixtureWithLocation.Location.CollectedAt = paymentFixtureWithLocation.Location.CollectedAt.Round(0)

	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyWithLocationFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixtureWithLocation)

	suite.Nil(response)
	suite.EqualError(err, ErrMissingLocationLatLong.Error())
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLogin() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterWebLogin() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginWebRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterWebLogin(loginWebFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWithEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, queryStringWithTrueEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithShouldEval)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterWebLoginWithEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginWebRequestBodyFixture, transactionAssessmentFixture, queryStringWithTrueEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterWebLogin(loginWebFixtureWithShouldEval)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWithFalseEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, queryStringWithFalseEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithShouldNotEval)
	suite.NoError(err)
	suite.Equal(emptyTransactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterWebLoginWithFalseEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginWebRequestBodyFixture, transactionAssessmentFixture, queryStringWithFalseEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterWebLogin(loginWebFixtureWithShouldNotEval)
	suite.NoError(err)
	suite.Equal(emptyTransactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWeb() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginWebRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.registerWebLogin(loginWebFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginAfterTokenExpiration() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.RegisterLogin(loginFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterWebLoginAfterTokenExpiration() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginWebRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterWebLogin(loginWebFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.RegisterWebLogin(loginWebFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterLoginNilLogin() {
	response, err := suite.client.RegisterLogin(nil)
	suite.EqualError(err, ErrMissingLogin.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterLoginNullInstallationIdAndSessionToken() {
	response, err := suite.client.RegisterLogin(&Login{AccountID: "some-account-id"})
	suite.EqualError(err, ErrMissingIdentifier.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterLoginEmptyAccountId() {
	response, err := suite.client.RegisterLogin(&Login{InstallationID: &installationId})
	suite.EqualError(err, ErrMissingAccountID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterLogin() {
	transactionServer := suite.mockPostTransactionsEndpoint("some-other-token", postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestUnauthorizedTokenGeneration() {
	tokenServer := suite.mockTokenEndpointUnauthorized()
	suite.client.tokenProvider.(*AutoRefreshTokenProvider).tokenClient.tokenEndpoint = tokenServer.URL
	defer tokenServer.Close()

	responsePayment, err := suite.client.RegisterPayment(paymentFixture)
	suite.Nil(responsePayment)
	suite.EqualError(err, ErrInvalidCredentials.Error())

	responseLogin, err := suite.client.RegisterLogin(loginFixture)
	suite.Nil(responseLogin)
	suite.EqualError(err, ErrInvalidCredentials.Error())

	responseSignUp, err := suite.client.RegisterSignup(installationId, addressFixture)
	suite.Nil(responseSignUp)
	suite.EqualError(err, ErrInvalidCredentials.Error())

	err = suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.EqualError(err, ErrInvalidCredentials.Error())
}

func (suite *IncogniaTestSuite) TestRegisterLoginErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Transactions = statusServer.URL

		response, err := suite.client.RegisterLogin(loginFixture)
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWithLocationAndTimestamp() {
	loginFixtureWithLocation.Location = locationFixtureFull
	postLoginRequestBodyWithLocationFixture.Location = locationFixtureFull

	*postLoginRequestBodyWithLocationFixture.Location.CollectedAt = postLoginRequestBodyWithLocationFixture.Location.CollectedAt.Round(0)
	*loginFixtureWithLocation.Location.CollectedAt = loginFixtureWithLocation.Location.CollectedAt.Round(0)

	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyWithLocationFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithLocation)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWithLocationWithoutTimestamp() {
	loginFixtureWithLocation.Location = locationFixtureMissingCollectedAt
	postLoginRequestBodyWithLocationFixture.Location = locationFixtureMissingCollectedAt

	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyWithLocationFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithLocation)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterLoginWithLocationMissingLat() {
	loginFixtureWithLocation.Location = locationFixtureMissingLat
	postLoginRequestBodyWithLocationFixture.Location = locationFixtureMissingLat

	*postLoginRequestBodyWithLocationFixture.Location.CollectedAt = postLoginRequestBodyWithLocationFixture.Location.CollectedAt.Round(0)
	*loginFixtureWithLocation.Location.CollectedAt = loginFixtureWithLocation.Location.CollectedAt.Round(0)

	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyWithLocationFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithLocation)

	suite.Nil(response)
	suite.EqualError(err, ErrMissingLocationLatLong.Error())
}

func (suite *IncogniaTestSuite) TestPanic() {
	defer func() { suite.Nil(recover()) }()

	panicString := "error getting token"
	suite.client.tokenProvider = &PanickingTokenProvider{panicString: panicString}

	suite.client.RegisterLogin(loginFixture)
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.Equal(err.Error(), panicString)
	_, err = suite.client.RegisterSignup("some-installationId", addressFixture)
	suite.Equal(err.Error(), panicString)
	_, err = suite.client.RegisterPayment(paymentFixture)
	suite.Equal(err.Error(), panicString)
}

func TestIncogniaTestSuite(t *testing.T) {
	suite.Run(t, new(IncogniaTestSuite))
}

func (suite *IncogniaTestSuite) mockFeedbackEndpoint(expectedToken string, expectedBody *postFeedbackRequestBody) *httptest.Server {
	feedbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		userAgent := r.Header.Get("User-Agent")
		suite.True(userAgentRegex.MatchString(userAgent), "User-Agent header does not match the expected format")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var requestBody postFeedbackRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if postFeedbackRequestBodyEqual(&requestBody, expectedBody) {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))

	suite.client.endpoints.Feedback = feedbackServer.URL

	return feedbackServer
}

func mockStatusServer(statusCode int) *httptest.Server {
	statusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(statusCode)
	}))

	return statusServer
}

func (suite *IncogniaTestSuite) mockTokenEndpointUnauthorized() *httptest.Server {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		userAgent := r.Header.Get("User-Agent")
		suite.True(userAgentRegex.MatchString(userAgent), "User-Agent header does not match the expected format")
	}))

	return tokenServer
}

func (suite *IncogniaTestSuite) mockPostTransactionsEndpoint(expectedToken string, expectedBody *postTransactionRequestBody, expectedResponse *TransactionAssessment, expectedQueryString map[string][]string) *httptest.Server {
	transactionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		userAgent := r.Header.Get("User-Agent")
		suite.True(userAgentRegex.MatchString(userAgent), "User-Agent header does not match the expected format")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		requestQueryString := r.URL.Query()
		for parameter := range requestQueryString {
			suite.Equal(expectedQueryString[parameter], requestQueryString[parameter])
		}

		requestEvalQueryString := requestQueryString["eval"]
		if requestEvalQueryString != nil && requestEvalQueryString[0] == "false" {
			res, _ := json.Marshal(emptyTransactionAssessmentFixture)
			w.Write(res)
			return
		}

		var requestBody postTransactionRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if reflect.DeepEqual(&requestBody, expectedBody) {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))

	suite.client.endpoints.Transactions = transactionsServer.URL

	return transactionsServer
}

func (suite *IncogniaTestSuite) mockPostSignupsEndpoint(expectedToken string, expectedBody *postAssessmentRequestBody, expectedResponse *SignupAssessment) *httptest.Server {
	signupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		userAgent := r.Header.Get("User-Agent")
		suite.True(userAgentRegex.MatchString(userAgent), "User-Agent header does not match the expected format")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var requestBody postAssessmentRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if reflect.DeepEqual(&requestBody, expectedBody) {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))

	suite.client.endpoints.Signups = signupsServer.URL

	return signupsServer
}

func isRequestAuthorized(request *http.Request, expectedToken string) bool {
	tokenType, token := readAuthorizationHeader(request)

	return token == expectedToken && tokenType == "Bearer"
}

func readAuthorizationHeader(request *http.Request) (string, string) {
	receivedToken := strings.Split(request.Header.Get("Authorization"), " ")
	tokenType := receivedToken[0]
	token := receivedToken[1]

	return tokenType, token
}

func mockTokenEndpoint(expectedToken string, expiresIn string) *httptest.Server {
	tokenResponse := map[string]string{
		"access_token": expectedToken,
		"expires_in":   expiresIn,
		"token_type":   "Bearer",
	}

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		username, password, ok := r.BasicAuth()

		if !ok || username != clientID || password != clientSecret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		res, _ := json.Marshal(tokenResponse)
		w.Write(res)
	}))

	return tokenServer
}

func postFeedbackRequestBodyEqual(a, b *postFeedbackRequestBody) bool {
	if a == nil || b == nil {
		return a == b
	}
	aOccurredAt := a.OccurredAt
	bOccurredAt := b.OccurredAt
	aCopy := *a
	aCopy.OccurredAt = nil
	bCopy := *b
	bCopy.OccurredAt = nil
	return reflect.DeepEqual(aCopy, bCopy) &&
		(aOccurredAt == nil && bOccurredAt == nil) ||
		(aOccurredAt != nil && bOccurredAt != nil && aOccurredAt.Equal(*bOccurredAt))
}
