package incognia

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const (
	defaultNetClientTimeout = 5 * time.Second
)

var (
	ErrMissingPayment                      = errors.New("missing payment parameters")
	ErrMissingLogin                        = errors.New("missing login parameters")
	ErrMissingSignup                       = errors.New("missing signup parameters")
	ErrMissingInstallationID               = errors.New("missing installation id")
	ErrMissingInstallationIDOrSessionToken = errors.New("missing installation id or session token")
	ErrMissingIdentifier                   = errors.New("missing installation id, request token or session token")
	ErrMissingAccountID                    = errors.New("missing account id")
	ErrMissingSignupID                     = errors.New("missing signup id")
	ErrMissingClientIDOrClientSecret       = errors.New("client id and client secret are required")
	ErrConfigIsNil                         = errors.New("incognia client config is required")
	ErrMissingLocationLatLong              = errors.New("location field missing latitude and/or longitude")
	ErrInvalidTimestamp                    = errors.New("location 'collected_at' attribute not in rfc3339 format")
)

type Client struct {
	clientID      string
	clientSecret  string
	tokenProvider TokenProvider
	netClient     *http.Client
	endpoints     *endpoints
	UserAgent     string
}

type IncogniaClientConfig struct {
	ClientID          string
	ClientSecret      string
	TokenProvider     TokenProvider
	Timeout           time.Duration
	TokenRouteTimeout time.Duration
	HTTPClient        *http.Client
}

type Payment struct {
	InstallationID   *string
	SessionToken     *string
	RequestToken     string
	AppVersion       string
	DeviceOs         string
	AccountID        string
	ExternalID       string
	PolicyID         string
	Location         *Location
	Coupon           *CouponType
	Addresses        []*TransactionAddress
	Value            *PaymentValue
	Methods          []*PaymentMethod
	Eval             *bool
	CustomProperties map[string]interface{}
}

type Login struct {
	InstallationID          *string
	SessionToken            *string
	RequestToken            string
	AccountID               string
	ExternalID              string
	PolicyID                string
	Location                *Location
	PaymentMethodIdentifier string
	Eval                    *bool
	AppVersion              string
	DeviceOs                string
	CustomProperties        map[string]interface{}
}

type FeedbackIdentifiers struct {
	InstallationID string
	SessionToken   string
	RequestToken   string
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

type Signup struct {
	InstallationID string
	RequestToken   string
	SessionToken   string
	AppVersion     string
	DeviceOs       string
	Address        *Address
	AccountID      string
	PolicyID       string
	ExternalID     string
}

func isRFC3339(timestamp string) bool {
	layout := time.RFC3339
	_, err := time.Parse(layout, timestamp)
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	return err == nil
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
	netClient := config.HTTPClient
	if netClient == nil {
		netClient = &http.Client{
			Timeout: timeout,
		}
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

	libVersion := "unknown"
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == "repo.incognia.com/go/incognia" {
				libVersion = dep.Version
			}
		}
	}

	userAgent := fmt.Sprintf(
		"incognia-api-go/%s (%s %s) Go/%s",
		libVersion,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)

	tokenProvider := config.TokenProvider
	if tokenProvider == nil {
		tokenProvider = NewAutoRefreshTokenProvider(tokenClient)
	}

	endpoints := getEndpoints()

	return &Client{clientID: config.ClientID, clientSecret: config.ClientSecret, tokenProvider: tokenProvider, netClient: netClient, endpoints: &endpoints, UserAgent: userAgent}, nil
}

func (c *Client) RegisterSignup(installationID string, address *Address) (ret *SignupAssessment, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			ret = nil
		}
	}()

	return c.registerSignup(&Signup{
		InstallationID: installationID,
		Address:        address,
	})
}

func (c *Client) RegisterSignupWithParams(params *Signup) (ret *SignupAssessment, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			ret = nil
		}
	}()

	return c.registerSignup(params)
}

func (c *Client) registerSignup(params *Signup) (ret *SignupAssessment, err error) {
	if params == nil {
		return nil, ErrMissingSignup
	}
	if params.InstallationID == "" && params.RequestToken == "" && params.SessionToken == "" {
		return nil, ErrMissingIdentifier
	}

	requestBody := postAssessmentRequestBody{
		InstallationID: params.InstallationID,
		RequestToken:   params.RequestToken,
		SessionToken:   params.SessionToken,
		AccountID:      params.AccountID,
		PolicyID:       params.PolicyID,
		ExternalID:     params.ExternalID,
		AppVersion:     params.AppVersion,
		DeviceOs:       strings.ToLower(params.DeviceOs),
	}
	if params.Address != nil {
		requestBody.AddressLine = params.Address.AddressLine
		requestBody.StructuredAddress = params.Address.StructuredAddress
		requestBody.Coordinates = params.Address.Coordinates
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

func (c *Client) RegisterFeedback(feedbackEvent FeedbackType, occurredAt *time.Time, feedbackIdentifiers *FeedbackIdentifiers) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return c.registerFeedback(feedbackEvent, occurredAt, nil, feedbackIdentifiers)
}

func (c *Client) RegisterFeedbackWithExpiration(feedbackEvent FeedbackType, occurredAt *time.Time, expiresAt *time.Time, feedbackIdentifiers *FeedbackIdentifiers) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return c.registerFeedback(feedbackEvent, occurredAt, expiresAt, feedbackIdentifiers)
}

func (c *Client) registerFeedback(feedbackEvent FeedbackType, occurredAt *time.Time, expiresAt *time.Time, feedbackIdentifiers *FeedbackIdentifiers) (err error) {
	requestBody := postFeedbackRequestBody{
		Event:      feedbackEvent,
		OccurredAt: occurredAt,
		ExpiresAt:  expiresAt,
	}
	if feedbackIdentifiers != nil {
		requestBody.InstallationID = feedbackIdentifiers.InstallationID
		requestBody.SessionToken = feedbackIdentifiers.SessionToken
		requestBody.RequestToken = feedbackIdentifiers.RequestToken
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

	if payment.InstallationID == nil && payment.SessionToken == nil && payment.RequestToken == "" {
		return nil, ErrMissingIdentifier
	}

	if payment.AccountID == "" {
		return nil, ErrMissingAccountID
	}

	if payment.Location != nil {
		if payment.Location.Latitude == nil || payment.Location.Longitude == nil {
			return nil, ErrMissingLocationLatLong
		}
		if payment.Location.Collected_at != "" && !isRFC3339(payment.Location.Collected_at) {
			return nil, ErrInvalidTimestamp
		}
	}

	requestBody, err := json.Marshal(postTransactionRequestBody{
		InstallationID:   payment.InstallationID,
		RequestToken:     payment.RequestToken,
		SessionToken:     payment.SessionToken,
		Type:             paymentType,
		AccountID:        payment.AccountID,
		PolicyID:         payment.PolicyID,
		Location:         payment.Location,
		Coupon:           payment.Coupon,
		ExternalID:       payment.ExternalID,
		Addresses:        payment.Addresses,
		PaymentValue:     payment.Value,
		PaymentMethods:   payment.Methods,
		AppVersion:       payment.AppVersion,
		DeviceOs:         strings.ToLower(payment.DeviceOs),
		CustomProperties: payment.CustomProperties,
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

	if login.InstallationID == nil && login.SessionToken == nil && login.RequestToken == "" {
		return nil, ErrMissingIdentifier
	}

	if login.AccountID == "" {
		return nil, ErrMissingAccountID
	}

	if login.Location != nil {
		if login.Location.Latitude == nil || login.Location.Longitude == nil {
			return nil, ErrMissingLocationLatLong
		}
		if login.Location.Collected_at != "" && !isRFC3339(login.Location.Collected_at) {
			return nil, ErrInvalidTimestamp
		}
	}

	requestBody, err := json.Marshal(postTransactionRequestBody{
		InstallationID:          login.InstallationID,
		Type:                    loginType,
		AccountID:               login.AccountID,
		PolicyID:                login.PolicyID,
		Location:                login.Location,
		ExternalID:              login.ExternalID,
		PaymentMethodIdentifier: login.PaymentMethodIdentifier,
		SessionToken:            login.SessionToken,
		RequestToken:            login.RequestToken,
		AppVersion:              login.AppVersion,
		DeviceOs:                strings.ToLower(login.DeviceOs),
		CustomProperties:        login.CustomProperties,
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
	request.Header.Add("User-Agent", c.UserAgent)

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
