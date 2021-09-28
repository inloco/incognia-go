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

func TestSuccessGetOnboardingAssessment(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupId := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupId, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.GetOnboardingAssessment(signupId)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}

func TestSuccessGetOnboardingAssessmentAfterTokenExpiration(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupId := "signup-id"
	signupServer := mockGetSignupsEndpoint(token, signupId, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.GetOnboardingAssessment(signupId)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)

	client.tokenManager.getToken().ExpiresIn = 0

	response, err = client.GetOnboardingAssessment(signupId)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}
func TestGetOnboardingAssessmentEmptySignupId(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	response, err := client.GetOnboardingAssessment("")
	assert.Error(t, err)
	assert.EqualError(t, err, "no signupId provided")
	assert.Nil(t, response)
}

func TestForbiddenGetOnboardingAssessment(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupId := "signup-id"
	signupServer := mockGetSignupsEndpoint("some-other-token", signupId, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.GetOnboardingAssessment(signupId)
	assert.Nil(t, response)
	assert.EqualError(t, err, "403 Forbidden")
}

func TestGetOnboardingAssessmentErrors(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		signupsEndpoint = statusServer.URL

		response, err := client.GetOnboardingAssessment("any-signup-id")
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), strconv.Itoa(status))
	}
}

func TestSuccessRegisterOnboardingAssessment(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.RegisterOnboardingAssessment(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}

func TestSuccessRegisterOnboardingAssessmentAfterTokenExpiration(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupServer := mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.RegisterOnboardingAssessment(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)

	client.tokenManager.getToken().ExpiresIn = 0

	response, err = client.RegisterOnboardingAssessment(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.NoError(t, err)
	assert.Equal(t, signupAssessmentFixture, response)
}
func TestRegisterOnboardingAssessmentEmptyInstallationId(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	response, err := client.RegisterOnboardingAssessment("", &Address{})
	assert.Error(t, err)
	assert.EqualError(t, err, "no installationId provided")
	assert.Nil(t, response)
}

func TestForbiddenRegisterOnboardingAssessment(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	signupServer := mockPostSignupsEndpoint("some-other-token", postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := client.RegisterOnboardingAssessment(postSignupRequestBodyFixture.InstallationId, addressFixture)
	assert.Nil(t, response)
	assert.EqualError(t, err, "403 Forbidden")
}

func TestRegisterOnboardingAssessmentErrors(t *testing.T) {
	client, _ := New(&IncogniaClientConfig{clientId, clientSecret})

	token := "some-token"
	tokenExpiresIn := "500"
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		signupsEndpoint = statusServer.URL

		response, err := client.RegisterOnboardingAssessment("any-signup-id", &Address{})
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
