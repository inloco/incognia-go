package incognia

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	clientId     string
	clientSecret string
	tokenManager *clientCredentialsTokenManager
	netClient    *http.Client
}

type IncogniaClientConfig struct {
	ClientId     string
	ClientSecret string
}

func New(config *IncogniaClientConfig) (*Client, error) {
	if config.ClientId == "" || config.ClientSecret == "" {
		return nil, errors.New("client id and client secret are required")
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	tokenManager := newClientCredentialsTokenManager(config.ClientId, config.ClientSecret)

	client := &Client{config.ClientId, config.ClientSecret, tokenManager, netClient}

	return client, nil
}

func (client *Client) GetOnboardingAssessment(signupId string) (*SignupAssessment, error) {
	if signupId == "" {
		return nil, errors.New("no signupId provided")
	}

	req, err := http.NewRequest("GET", signupsEndpoint+"/"+signupId, nil)
	if err != nil {
		return nil, err
	}

	var signupAssessment SignupAssessment

	err = client.doRequest(req, &signupAssessment)
	if err != nil {
		return nil, err
	}

	return &signupAssessment, nil
}

func (client *Client) RegisterOnboardingAssessment(installationId string, address *Address) (*SignupAssessment, error) {
	if installationId == "" {
		return nil, errors.New("no installationId provided")
	}

	requestBody, err := json.Marshal(postAssessmentRequestBody{
		InstallationId:    installationId,
		AddressLine:       address.AddressLine,
		StructuredAddress: address.StructuredAddress,
		Coordinates:       address.Coordinates,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", signupsEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	var signupAssessment SignupAssessment

	req.Header.Add("Content-Type", "application/json")

	err = client.doRequest(req, &signupAssessment)
	if err != nil {
		return nil, err
	}

	return &signupAssessment, nil
}

func (client *Client) doRequest(request *http.Request, response interface{}) error {
	client.authorizeRequest(request)

	res, err := client.netClient.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if len(body) > 0 {
			return errors.New(res.Status + " " + string(body))
		}

		return errors.New(res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) authorizeRequest(request *http.Request) {
	token := client.tokenManager.getToken()
	request.Header.Add("Authorization", token.TokenType+" "+token.AccessToken)
}
