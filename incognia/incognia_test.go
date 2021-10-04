package incognia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

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
		Id:             "some-id",
		DeviceId:       "some-device-id",
		RequestId:      "some-request-id",
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
		InstallationId: "installation-id",
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

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessment() {
	defer suite.tokenServer.Close()

	signupId := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupId, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupId)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessmentAfterTokenExpiration() {
	defer suite.tokenServer.Close()

	signupId := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupId, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupId)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenManager.getToken()
	token.ExpiresIn = 0

	response, err = suite.client.GetSignupAssessment(signupId)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}
func (suite *IncogniaTestSuite) TestGetSignupAssessmentEmptySignupId() {
	defer suite.tokenServer.Close()

	response, err := suite.client.GetSignupAssessment("")
	suite.EqualError(err, "no signupID provided")
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenGetSignupAssessment() {
	defer suite.tokenServer.Close()

	signupId := "signup-id"
	signupServer := mockGetSignupsEndpoint("some-other-token", signupId, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupId)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestGetSignupAssessmentErrors() {
	defer suite.tokenServer.Close()

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
	defer suite.tokenServer.Close()

	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupAfterTokenExpiration() {
	defer suite.tokenServer.Close()

	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenManager.getToken()
	token.ExpiresIn = 0

	response, err = suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}
func (suite *IncogniaTestSuite) TestRegisterSignupEmptyInstallationId() {
	defer suite.tokenServer.Close()

	response, err := suite.client.RegisterSignup("", &Address{})
	suite.EqualError(err, "no installationId provided")
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterSignup() {
	defer suite.tokenServer.Close()

	signupServer := mockPostSignupsEndpoint("some-other-token", postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestRegisterSignupErrors() {
	defer suite.tokenServer.Close()

	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		signupsEndpoint = statusServer.URL

		response, err := suite.client.RegisterSignup("any-signup-id", &Address{})
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func TestIncogniaTestSuite(t *testing.T) {
	suite.Run(t, new(IncogniaTestSuite))
}

func mockStatusServer(statusCode int) *httptest.Server {
	statusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(statusCode)
	}))

	return statusServer
}

func mockPostSignupsEndpoint(expectedToken string, expectedBody *postAssessmentRequestBody, expectedResponse *SignupAssessment) *httptest.Server {
	signupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		tokenType, token := readAuthorizationHeader(r)
		if token != expectedToken || tokenType != "Bearer" {
			w.WriteHeader(403)
			return
		}

		var requestBody postAssessmentRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if reflect.DeepEqual(&requestBody, expectedBody) {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	signupsEndpoint = signupsServer.URL

	return signupsServer
}

func mockGetSignupsEndpoint(expectedToken, expectedSignupId string, expectedResponse *SignupAssessment) *httptest.Server {
	getSignupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		tokenType, token := readAuthorizationHeader(r)

		if token != expectedToken || tokenType != "Bearer" {
			w.WriteHeader(403)
			return
		}

		defer r.Body.Close()

		splitUrl := strings.Split(r.URL.RequestURI(), "/")
		requestSignupId := splitUrl[len(splitUrl)-1]

		if requestSignupId == expectedSignupId {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	signupsEndpoint = getSignupsServer.URL

	return getSignupsServer
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
			w.WriteHeader(401)
			return
		}

		res, _ := json.Marshal(tokenResponse)
		w.Write(res)
	}))

	tokenEndpoint = tokenServer.URL

	return tokenServer
}
