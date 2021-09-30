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

func (c *Client) RegisterSignup(installationID string, address *Address) (*SignupAssessment, error) {
	if installationID == "" {
		return nil, errors.New("no installationId provided")
	}

	requestBody, err := json.Marshal(postAssessmentRequestBody{
		InstallationID:    installationID,
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

	err = c.doRequest(req, &signupAssessment)
	if err != nil {
		return nil, err
	}

	return &signupAssessment, nil
}

func (client *Client) RegisterFeedback(feedbackEvent FeedbackType, timestamp *time.Time, feedbackIdentifiers *FeedbackIdentifiers) error {
	requestBody, err := json.Marshal(postFeedbackRequestBody{
		Event:          feedbackEvent,
		Timestamp:      timestamp.UnixMilli(),
		InstallationId: feedbackIdentifiers.InstallationId,
		LoginId:        feedbackIdentifiers.LoginId,
		PaymentId:      feedbackIdentifiers.PaymentId,
		SignupId:       feedbackIdentifiers.SignupId,
		AccountId:      feedbackIdentifiers.AccountId,
		ExternalId:     feedbackIdentifiers.ExternalId,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", feedbackEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	err = client.doRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) doRequest(request *http.Request, response interface{}) error {
	request.Header.Add("Content-Type", "application/json")

	err := c.authorizeRequest(request)
	if err != nil {
		return err
	}

	res, err := c.netClient.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		if len(body) > 0 {
			return errors.New(res.Status + " " + string(body))
		}

		return errors.New(res.Status)
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &response)
		if err != nil {
			return err
		}
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
