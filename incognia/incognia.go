package incognia

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	clientID     string
	clientSecret string
	tokenManager *clientCredentialsTokenManager
	netClient    *http.Client
}

type IncogniaClientConfig struct {
	ClientID     string
	ClientSecret string
}

func New(config *IncogniaClientConfig) (*Client, error) {
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, errors.New("client id and client secret are required")
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	tokenManager := newClientCredentialsTokenManager(config.ClientID, config.ClientSecret)

	client := &Client{config.ClientID, config.ClientSecret, tokenManager, netClient}

	return client, nil
}

func (c *Client) GetSignupAssessment(signupID string) (*SignupAssessment, error) {
	if signupID == "" {
		return nil, errors.New("no signupID provided")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", signupsEndpoint, signupID), nil)
	if err != nil {
		return nil, err
	}

	var signupAssessment SignupAssessment

	err = c.doRequest(req, &signupAssessment)
	if err != nil {
		return nil, err
	}

	return &signupAssessment, nil
}

func (c *Client) RegisterSignup(installationId string, address *Address) (*SignupAssessment, error) {
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

	err = c.doRequest(req, &signupAssessment)
	if err != nil {
		return nil, err
	}

	return &signupAssessment, nil
}

func (c *Client) doRequest(request *http.Request, response interface{}) error {
	err := c.authorizeRequest(request)
	if err != nil {
		return err
	}

	res, err := c.netClient.Do(request)
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

func (c *Client) authorizeRequest(request *http.Request) error {
	token, err := c.tokenManager.getToken()
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", token.TokenType+" "+token.AccessToken)

	return nil
}
