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

const (
	defaultNetClientTimeout = 5 * time.Second
)

var (
	ErrMissingPayment                      = errors.New("missing payment parameters")
	ErrMissingLogin                        = errors.New("missing login parameters")
	ErrMissingInstallationID               = errors.New("missing installation id")
	ErrMissingInstallationIDOrSessionToken = errors.New("missing installation id or session token")
	ErrMissingAccountID                    = errors.New("missing account id")
	ErrMissingSignupID                     = errors.New("missing signup id")
	ErrMissingTimestamp                    = errors.New("missing timestamp")
	ErrInvalidFeedbackType                 = errors.New("invalid feedback type")
	ErrMissingClientIDOrClientSecret       = errors.New("client id and client secret are required")
	ErrConfigIsNil                         = errors.New("incognia client config is required")
)

type Region int64

const (
	US Region = iota
	BR
)

type Client struct {
	clientID      string
	clientSecret  string
	tokenProvider TokenProvider
	netClient     *http.Client
	endpoints     *endpoints
}

type IncogniaClientConfig struct {
	ClientID          string
	ClientSecret      string
	TokenProvider     TokenProvider
	Timeout           time.Duration
	TokenRouteTimeout time.Duration
	// Deprecated: Region is no longer used to determine endpoints
	Region Region
}

type Payment struct {
	InstallationID string
	AccountID      string
	ExternalID     string
	PolicyID       string
	Addresses      []*TransactionAddress
	Value          *PaymentValue
	Methods        []*PaymentMethod
	Eval           *bool
}

type Login struct {
	InstallationID          *string
	SessionToken            *string
	AccountID               string
	ExternalID              string
	PolicyID                string
	PaymentMethodIdentifier string
	Eval                    *bool
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
	if config == nil {
		return nil, ErrConfigIsNil
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, ErrMissingClientIDOrClientSecret
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultNetClientTimeout
	}
	netClient := &http.Client{
		Timeout: timeout,
	}

	tokenRouteTimeout := config.TokenRouteTimeout
	if tokenRouteTimeout == 0 {
		tokenRouteTimeout = defaultNetClientTimeout
	}

	tokenClient := NewTokenClient(&TokenClientConfig{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Timeout:      tokenRouteTimeout,
	})

	tokenProvider := config.TokenProvider
	if tokenProvider == nil {
		tokenProvider = NewAutoRefreshTokenProvider(tokenClient)
	}

	endpoints := getEndpoints()

	return &Client{clientID: config.ClientID, clientSecret: config.ClientSecret, tokenProvider: tokenProvider, netClient: netClient, endpoints: &endpoints}, nil
}

func (c *Client) GetSignupAssessment(signupID string) (ret *SignupAssessment, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			ret = nil
		}
	}()

	return c.getSignupAssessment(signupID)
}

func (c *Client) getSignupAssessment(signupID string) (ret *SignupAssessment, err error) {
	if signupID == "" {
		return nil, ErrMissingSignupID
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

func (c *Client) RegisterSignup(installationID string, address *Address) (ret *SignupAssessment, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			ret = nil
		}
	}()

	return c.registerSignup(installationID, address)
}

func (c *Client) registerSignup(installationID string, address *Address) (ret *SignupAssessment, err error) {
	if installationID == "" {
		return nil, ErrMissingInstallationID
	}

	requestBody := postAssessmentRequestBody{
		InstallationID: installationID,
	}
	if address != nil {
		requestBody.AddressLine = address.AddressLine
		requestBody.StructuredAddress = address.StructuredAddress
		requestBody.Coordinates = address.Coordinates
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.endpoints.Signups, bytes.NewBuffer(requestBodyBytes))
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

func (c *Client) RegisterFeedback(feedbackEvent FeedbackType, timestamp *time.Time, feedbackIdentifiers *FeedbackIdentifiers) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return c.registerFeedback(feedbackEvent, timestamp, feedbackIdentifiers)
}

func (c *Client) registerFeedback(feedbackEvent FeedbackType, timestamp *time.Time, feedbackIdentifiers *FeedbackIdentifiers) (err error) {
	if !isValidFeedbackType(feedbackEvent) {
		return ErrInvalidFeedbackType
	}
	if timestamp == nil {
		return ErrMissingTimestamp
	}

	requestBody := postFeedbackRequestBody{
		Event:     feedbackEvent,
		Timestamp: timestamp.UnixNano() / 1000000,
	}
	if feedbackIdentifiers != nil {
		requestBody.InstallationID = feedbackIdentifiers.InstallationID
		requestBody.LoginID = feedbackIdentifiers.LoginID
		requestBody.PaymentID = feedbackIdentifiers.PaymentID
		requestBody.SignupID = feedbackIdentifiers.SignupID
		requestBody.AccountID = feedbackIdentifiers.AccountID
		requestBody.ExternalID = feedbackIdentifiers.ExternalID
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoints.Feedback, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return err
	}

	err = c.doRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RegisterPayment(payment *Payment) (ret *TransactionAssessment, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			ret = nil
		}
	}()

	return c.registerPayment(payment)
}

func (c *Client) registerPayment(payment *Payment) (ret *TransactionAssessment, err error) {
	if payment == nil {
		return nil, ErrMissingPayment
	}

	if payment.InstallationID == "" {
		return nil, ErrMissingInstallationID
	}

	if payment.AccountID == "" {
		return nil, ErrMissingAccountID
	}

	requestBody, err := json.Marshal(postTransactionRequestBody{
		InstallationID: &payment.InstallationID,
		Type:           paymentType,
		AccountID:      payment.AccountID,
		PolicyID:       payment.PolicyID,
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

	if payment.Eval != nil {
		q := req.URL.Query()
		q.Add("eval", fmt.Sprintf("%t", *payment.Eval))
		req.URL.RawQuery = q.Encode()
	}

	var paymentAssesment TransactionAssessment

	err = c.doRequest(req, &paymentAssesment)
	if err != nil {
		return nil, err
	}

	return &paymentAssesment, nil
}

func (c *Client) RegisterLogin(login *Login) (ret *TransactionAssessment, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			ret = nil
		}
	}()

	return c.registerLogin(login)
}

func (c *Client) registerLogin(login *Login) (*TransactionAssessment, error) {
	if login == nil {
		return nil, ErrMissingLogin
	}

	if login.InstallationID == nil && login.SessionToken == nil {
		return nil, ErrMissingInstallationIDOrSessionToken
	}

	if login.AccountID == "" {
		return nil, ErrMissingAccountID
	}

	requestBody, err := json.Marshal(postTransactionRequestBody{
		InstallationID:          login.InstallationID,
		Type:                    loginType,
		AccountID:               login.AccountID,
		PolicyID:                login.PolicyID,
		ExternalID:              login.ExternalID,
		PaymentMethodIdentifier: login.PaymentMethodIdentifier,
		SessionToken:            login.SessionToken,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.endpoints.Transactions, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	if login.Eval != nil {
		q := req.URL.Query()
		q.Add("eval", fmt.Sprintf("%t", *login.Eval))
		req.URL.RawQuery = q.Encode()
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
	token, err := c.tokenProvider.GetToken()
	if err != nil {
		return err
	}

	token.SetAuthHeader(request)

	return nil
}

func isValidFeedbackType(feedbackType FeedbackType) bool {
	switch feedbackType {
	case
		PaymentAccepted,
		PaymentDeclined,
		PaymentDeclinedByRiskAnalysis,
		PaymentDeclinedByAcquirer,
		PaymentDeclinedByBusiness,
		PaymentDeclinedByManualReview,
		PaymentAcceptedByThirdParty,
		LoginAccepted,
		LoginDeclined,
		SignupAccepted,
		SignupDeclined,
		ChallengePassed,
		ChallengeFailed,
		PasswordChangedSuccessfully,
		PasswordChangeFailed,
		Verified,
		NotVerified,
		Chargeback,
		PromotionAbuse,
		AccountTakeover,
		MposFraud,
		ChargebackNotification:
		return true
	default:
		return false
	}
}
