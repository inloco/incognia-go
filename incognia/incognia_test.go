package incognia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	clientId     string = "client-id"
	clientSecret string = "client-secret"
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

func TestSuccessGetSignupAssessment(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupID := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.GetSignupAssessment(signupID)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}

func TestSuccessGetSignupAssessmentAfterTokenExpiration(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupID := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.GetSignupAssessment(signupID)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)

	client.tokenManager.getToken().ExpiresIn = 0

	response, err = client.GetSignupAssessment(signupID)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}
func TestGetSignupAssessmentEmptysignupID(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	response, err := client.GetSignupAssessment("")
	assert.EqualError(t, err, "no signupID provided")
	assert.Nil(t, response)
}

func TestForbiddenGetSignupAssessment(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupID := "signup-id"
	signupServer := mockGetSignupsEndpoint("some-other-token", signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.GetSignupAssessment(signupID)
	assert.Nil(t, response)
	assert.EqualError(t, err, "403 Forbidden")
}

func TestGetSignupAssessmentErrors(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		signupsEndpoint = statusServer.URL

		response, err := client.GetSignupAssessment("any-signup-id")
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), strconv.Itoa(status))
	}
}

func TestSuccessRegisterSignup(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}

func TestSuccessRegisterSignupAfterTokenExpiration(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)

	client.tokenManager.getToken().ExpiresIn = 0

	response, err = client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}
func TestRegisterSignupEmptyInstallationId(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	response, err := client.RegisterSignup("", &Address{})
	assert.EqualError(t, err, "no installationId provided")
	assert.Nil(t, response)
}

func TestForbiddenRegisterSignup(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupServer := mockPostSignupsEndpoint("some-other-token", postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.RegisterSignup(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.Nil(t, response)
	assert.EqualError(t, err, "403 Forbidden")
}

func TestRegisterSignupErrors(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		signupsEndpoint = statusServer.URL

		response, err := client.RegisterSignup("any-signup-id", &Address{})
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), strconv.Itoa(status))
	}
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

func mockGetSignupsEndpoint(expectedToken, expectedsignupID string, expectedResponse *SignupAssessment) *httptest.Server {
	getSignupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		tokenType, token := readAuthorizationHeader(r)

		if token != expectedToken || tokenType != "Bearer" {
			w.WriteHeader(403)
			return
		}

		defer r.Body.Close()

		splitUrl := strings.Split(r.URL.RequestURI(), "/")
		requestsignupID := splitUrl[len(splitUrl)-1]

		if requestsignupID == expectedsignupID {
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

		if !ok || username != clientId || password != clientSecret {
			w.WriteHeader(401)
			return
		}

		res, _ := json.Marshal(tokenResponse)
		w.Write(res)
	}))

	tokenEndpoint = tokenServer.URL

	return tokenServer
}
