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
	userAgentRegex  = regexp.MustCompile(`^incognia-api-go(/(v[0-9]+\.[0-9]+\.[0-9]+|unknown))? \([a-z]+ [a-z0-9]+\) Go/go[0-9]+\.[0-9]+\.[0-9]+$`)
	now             = time.Now()
	nowMinusSeconds = now.Add(-1 * time.Second)
	installationId  = "installation-id"
	requestToken    = "request-token"
	customProperty  = map[string]interface{}{
		"custom_1": "custom_value_1",
		"custom_2": "custom_value_2",
	}
	shouldEval               bool                = true
	shouldNotEval            bool                = false
	emptyQueryString         map[string][]string = nil
	queryStringWithFalseEval                     = map[string][]string{"eval": []string{"false"}}
	queryStringWithTrueEval                      = map[string][]string{"eval": []string{"true"}}
	signupAssessmentFixture                      = &SignupAssessment{
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
		InstallationID: installationId,
		RequestToken:   requestToken,
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
		AccountID:  "account-id",
		PolicyID:   "policy-id",
		ExternalID: "external-id",
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
		PolicyID:       "policy-id",
		Type:           paymentType,
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
		PaymentValue: &PaymentValue{
			Amount:   55.02,
			Currency: "BRL",
		},
		PaymentMethods: []*PaymentMethod{
			{
				Type:       CreditCard,
				Identifier: "credit-card-hash-123",
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
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
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
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
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
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
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
	}
	paymentFixtureRequiredFields = &Payment{
		InstallationID: &installationId,
		AccountID:      "account-id",
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
		CustomProperties:        customProperty,
		PaymentMethodIdentifier: "payment-method-identifier",
	}
	loginFixtureWithShouldEval = &Login{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		Eval:                    &shouldEval,
		CustomProperties:        customProperty,
	}
	loginFixtureWithShouldNotEval = &Login{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Eval:           &shouldNotEval,
	}
	loginWebFixture = &Login{
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		RequestToken:            requestToken,
	}
	postLoginRequestBodyFixture = &postTransactionRequestBody{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		Type:                    loginType,
		CustomProperties:        customProperty,
	}
	postLoginWebRequestBodyFixture = &postTransactionRequestBody{
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		Type:                    loginType,
		RequestToken:            requestToken,
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

	_, err := client.GetSignupAssessment("any-signup-id")
	suite.EqualError(err, ErrTokenNotFound.Error())

	_, err = client.RegisterLogin(loginFixture)
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
	signupID := "signup-id"

	signupServer := suite.mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()
	_, err := client.GetSignupAssessment(signupID)
	suite.NoError(err)

	loginServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer loginServer.Close()
	_, err = client.RegisterLogin(loginFixture)
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

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessment() {
	signupID := "signup-id"
	signupServer := suite.mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessmentAfterTokenExpiration() {
	signupID := "signup-id"
	signupServer := suite.mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}
func (suite *IncogniaTestSuite) TestGetSignupAssessmentEmptySignupId() {
	response, err := suite.client.GetSignupAssessment("")
	suite.EqualError(err, ErrMissingSignupID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenGetSignupAssessment() {
	signupID := "signup-id"
	signupServer := suite.mockGetSignupsEndpoint("some-other-token", signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestGetSignupAssessmentErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Signups = statusServer.URL

		response, err := suite.client.GetSignupAssessment("any-signup-id")
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupWithParams() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyWithAllParamsFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignupWithParams(&Signup{
		InstallationID: postSignupRequestBodyWithAllParamsFixture.InstallationID,
		RequestToken:   postSignupRequestBodyWithAllParamsFixture.RequestToken,
		SessionToken:   postSignupRequestBodyWithAllParamsFixture.SessionToken,
		Address:        addressFixture,
		AccountID:      postSignupRequestBodyWithAllParamsFixture.AccountID,
		PolicyID:       postSignupRequestBodyWithAllParamsFixture.PolicyID,
		ExternalID:     postSignupRequestBodyWithAllParamsFixture.ExternalID,
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

func (suite *IncogniaTestSuite) TestSuccessRegisterLogin() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixture)
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

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWithFalseEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, queryStringWithFalseEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithShouldNotEval)
	suite.NoError(err)
	suite.Equal(emptyTransactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWeb() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginWebRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.registerLogin(loginWebFixture)
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

func (suite *IncogniaTestSuite) TestPanic() {
	defer func() { suite.Nil(recover()) }()

	panicString := "error getting token"
	suite.client.tokenProvider = &PanickingTokenProvider{panicString: panicString}

	suite.client.RegisterLogin(loginFixture)
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, postFeedbackRequestBodyFixture.OccurredAt, feedbackIdentifiersFixture)
	suite.Equal(err.Error(), panicString)
	_, err = suite.client.RegisterSignup("some-installationId", addressFixture)
	suite.Equal(err.Error(), panicString)
	_, err = suite.client.GetSignupAssessment("some-signup-id")
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

func (suite *IncogniaTestSuite) mockGetSignupsEndpoint(expectedToken, expectedSignupID string, expectedResponse *SignupAssessment) *httptest.Server {
	getSignupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		defer r.Body.Close()

		splitUrl := strings.Split(r.URL.RequestURI(), "/")
		requestSignupID := splitUrl[len(splitUrl)-1]

		if requestSignupID == expectedSignupID {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	suite.client.endpoints.Signups = getSignupsServer.URL

	return getSignupsServer
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
