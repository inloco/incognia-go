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

type Region int64

const (
	US Region = iota
	BR
)

type Client struct {
	clientID     string
	clientSecret string
	tokenManager *clientCredentialsTokenManager
	netClient    *http.Client
	endpoints    *endpoints
}

type IncogniaClientConfig struct {
	ClientID     string
	ClientSecret string
	Region       Region
}

type Payment struct {
	InstallationID string
	AccountID      string
	ExternalID     string
	Addresses      []*TransactionAddress
	Value          *PaymentValue
	Methods        []*PaymentMethod
}

type Login struct {
	InstallationID string
	AccountID      string
	ExternalID     string
}

type FeedbackIdentifiers struct {
	InstallationID string
	LoginID        string
	PaymentID      string
	SignupID       string
	AccountID      string
	ExternalID     string
}

type Address struct {
	Coordinates       *Coordinates
	StructuredAddress *StructuredAddress
	AddressLine       string
}

func New(config *IncogniaClientConfig) (*Client, error) {
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, errors.New("client id and client secret are required")
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	endpoints := newEndpoints(config.Region)

	tokenManager := newClientCredentialsTokenManager(&clientCredentialsManagerConfig{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     endpoints.Token,
		NetClient:    netClient,
	})

	client := &Client{config.ClientID, config.ClientSecret, tokenManager, netClient, &endpoints}

	return client, nil
}

func (c *Client) GetSignupAssessment(signupID string) (*SignupAssessment, error) {
	if signupID == "" {
		return nil, errors.New("no signupID provided")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", c.endpoints.Signups, signupID), nil)
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

	req, err := http.NewRequest("POST", c.endpoints.Signups, bytes.NewBuffer(requestBody))
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

func (c *Client) RegisterFeedback(feedbackEvent FeedbackType, timestamp *time.Time, feedbackIdentifiers *FeedbackIdentifiers) error {
	requestBody, err := json.Marshal(postFeedbackRequestBody{
		Event:          feedbackEvent,
		Timestamp:      timestamp.UnixNano() / 1000000,
		InstallationID: feedbackIdentifiers.InstallationID,
		LoginID:        feedbackIdentifiers.LoginID,
		PaymentID:      feedbackIdentifiers.PaymentID,
		SignupID:       feedbackIdentifiers.SignupID,
		AccountID:      feedbackIdentifiers.AccountID,
		ExternalID:     feedbackIdentifiers.ExternalID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoints.Feedback, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	err = c.doRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RegisterPayment(payment *Payment) (*TransactionAssessment, error) {
	if payment.InstallationID == "" {
		return nil, errors.New("missing installation id")
	}

	if payment.AccountID == "" {
		return nil, errors.New("missing account id")
	}

	requestBody, err := json.Marshal(postTransactionRequestBody{
		InstallationID: payment.InstallationID,
		Type:           paymentType,
		AccountID:      payment.AccountID,
		ExternalID:     payment.ExternalID,
		Addresses:      payment.Addresses,
		PaymentValue:   payment.Value,
		PaymentMethods: payment.Methods,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.endpoints.Transactions, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	var paymentAssesment TransactionAssessment

	err = c.doRequest(req, &paymentAssesment)
	if err != nil {
		return nil, err
	}

	return &paymentAssesment, nil
}

func (c *Client) RegisterLogin(login *Login) (*TransactionAssessment, error) {
	if login.InstallationID == "" {
		return nil, errors.New("missing installation id")
	}

	if login.AccountID == "" {
		return nil, errors.New("missing account id")
	}

	requestBody, err := json.Marshal(postTransactionRequestBody{
		InstallationID: login.InstallationID,
		Type:           loginType,
		AccountID:      login.AccountID,
		ExternalID:     login.ExternalID,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.endpoints.Transactions, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	var loginAssessment TransactionAssessment

	err = c.doRequest(req, &loginAssessment)
	if err != nil {
		return nil, err
	}

	return &loginAssessment, nil
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
			return fmt.Errorf("%s %s", res.Status, string(body))
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

	request.Header.Add("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))

	return nil
}
