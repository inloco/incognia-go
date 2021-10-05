package incognia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
	signupAssessmentFixture = &SignupAssessment{
		ID:             "some-id",
		DeviceID:       "some-device-id",
		RequestID:      "some-request-id",
		RiskAssessment: LowRisk,
		Evidence: map[string]interface{}{
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
		InstallationID: "installation-id",
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
	addressFixture = &Address{
		Coordinates:       postSignupRequestBodyFixture.Coordinates,
		StructuredAddress: postSignupRequestBodyFixture.StructuredAddress,
		AddressLine:       postSignupRequestBodyFixture.AddressLine,
	}
	postFeedbackRequestBodyFixture = &postFeedbackRequestBody{
		Event:          SignupAccepted,
		Timestamp:      time.Now().UnixMilli(),
		InstallationID: "some-installation-id",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
	}
	feedbackIdentifiersFixture = &FeedbackIdentifiers{
		InstallationID: "some-installation-id",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
	}
	transactionAssessmentFixture = &TransactionAssessment{
		ID:             "some-id",
		DeviceID:       "some-device-id",
		RiskAssessment: LowRisk,
		Evidence: map[string]interface{}{
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
		InstallationID: "installation-id",
		AccountID:      "account-id",
		ExternalID:     "external-id",
		Type:           paymentType,
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
	}
	paymentFixture = &Payment{
		InstallationID: "installation-id",
		AccountID:      "account-id",
		ExternalID:     "external-id",
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
	}
)

type IncogniaTestSuite struct {
	suite.Suite

	client      *Client
	token       string
	tokenServer *httptest.Server
}

func (suite *IncogniaTestSuite) SetupTest() {
	client, _ := New(&IncogniaClientConfig{clientID, clientSecret})
	suite.client = client

	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	suite.token = token
	suite.tokenServer = tokenServer
}

func (suite *IncogniaTestSuite) TearDownTest() {
	defer suite.tokenServer.Close()
}

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessment() {
	signupID := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessmentAfterTokenExpiration() {
	signupID := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenManager.getToken()
	token.ExpiresIn = 0

	response, err = suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}
func (suite *IncogniaTestSuite) TestGetSignupAssessmentEmptySignupId() {
	response, err := suite.client.GetSignupAssessment("")
	suite.EqualError(err, "no signupID provided")
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenGetSignupAssessment() {
	signupID := "signup-id"
	signupServer := mockGetSignupsEndpoint("some-other-token", signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestGetSignupAssessmentErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		signupsEndpoint = statusServer.URL

		response, err := suite.client.GetSignupAssessment("any-signup-id")
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignup() {
	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupAfterTokenExpiration() {
	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenManager.getToken()
	token.ExpiresIn = 0

	response, err = suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}
func (suite *IncogniaTestSuite) TestRegisterSignupEmptyInstallationId() {
	response, err := suite.client.RegisterSignup("", &Address{})
	suite.EqualError(err, "no installationId provided")
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterSignup() {
	signupServer := mockPostSignupsEndpoint("some-other-token", postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestRegisterSignupErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		signupsEndpoint = statusServer.URL

		response, err := suite.client.RegisterSignup("any-signup-id", &Address{})
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedback() {
	feedbackServer := mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	timestamp := time.UnixMilli(postFeedbackRequestBodyFixture.Timestamp)
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackAfterTokenExpiration() {
	feedbackServer := mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	timestamp := time.UnixMilli(postFeedbackRequestBodyFixture.Timestamp)

	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.NoError(err)

	token, _ := suite.client.tokenManager.getToken()
	token.ExpiresIn = 0

	err = suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterFeedback() {
	feedbackServer := mockFeedbackEndpoint("some-other-token", postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	timestamp := time.UnixMilli(postFeedbackRequestBodyFixture.Timestamp)
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestErrorsRegisterFeedback() {
	timestamp := time.UnixMilli(postFeedbackRequestBodyFixture.Timestamp)

	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		feedbackEndpoint = statusServer.URL

		err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPayment() {
	transactionServer := mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentAfterTokenExpiration() {
	transactionServer := mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)

	token, _ := suite.client.tokenManager.getToken()
	token.ExpiresIn = 0

	response, err = suite.client.RegisterPayment(paymentFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}
func (suite *IncogniaTestSuite) TestRegisterPaymentEmptyInstallationId() {
	response, err := suite.client.RegisterPayment(&Payment{AccountID: "some-account-id"})
	suite.EqualError(err, "missing installation id")
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentEmptyAccountId() {
	response, err := suite.client.RegisterPayment(&Payment{InstallationID: "some-installation-id"})
	suite.EqualError(err, "missing account id")
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterPayment() {
	transactionServer := mockPostTransactionsEndpoint("some-other-token", postPaymentRequestBodyFixture, transactionAssessmentFixture)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestRegisterPaymentErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		transactionsEndpoint = statusServer.URL

		response, err := suite.client.RegisterPayment(paymentFixture)
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func TestIncogniaTestSuite(t *testing.T) {
	suite.Run(t, new(IncogniaTestSuite))
}

func mockFeedbackEndpoint(expectedToken string, expectedBody *postFeedbackRequestBody) *httptest.Server {
	feedbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var requestBody postFeedbackRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if reflect.DeepEqual(&requestBody, expectedBody) {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))

	feedbackEndpoint = feedbackServer.URL

	return feedbackServer
}

func mockStatusServer(statusCode int) *httptest.Server {
	statusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(statusCode)
	}))

	return statusServer
}

func mockPostTransactionsEndpoint(expectedToken string, expectedBody *postTransactionRequestBody, expectedResponse *TransactionAssessment) *httptest.Server {
	transactionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
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

	transactionsEndpoint = transactionsServer.URL

	return transactionsServer
}

func mockPostSignupsEndpoint(expectedToken string, expectedBody *postAssessmentRequestBody, expectedResponse *SignupAssessment) *httptest.Server {
	signupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

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

	signupsEndpoint = signupsServer.URL

	return signupsServer
}

func mockGetSignupsEndpoint(expectedToken, expectedSignupID string, expectedResponse *SignupAssessment) *httptest.Server {
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

	signupsEndpoint = getSignupsServer.URL

	return getSignupsServer
}

func isRequestAuthorized(request *http.Request, expectedToken string) bool {
	tokenType, token := readAuthorizationHeader(request)

	if token != expectedToken || tokenType != "Bearer" {
		return false
	}

	return true
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

	tokenEndpoint = tokenServer.URL

	return tokenServer
}
