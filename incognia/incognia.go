package incognia

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Client struct {
	clientId      string
	clientSecret  string
	endpoints     endpoints
	incogniaToken *incogniaToken
	netClient     *http.Client
}

type IncogniaClientConfig struct {
	ClientId     string
	ClientSecret string
	Region       Region
}

func New(config *IncogniaClientConfig) (*Client, error) {
	if config.ClientId == "" || config.ClientSecret == "" {
		return nil, errors.New("client id and client secret are required")
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	endpoints := buildEndpoints(config.Region)

	client := &Client{config.ClientId, config.ClientSecret, endpoints, nil, netClient}

	return client, nil
}

func (c *Client) refreshToken() error {
	req, _ := http.NewRequest("POST", c.endpoints.Token, nil)

	req.SetBasicAuth(c.clientId, c.clientSecret)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := c.netClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var result incogniaToken

	result.CreatedAt = time.Now().Unix()

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return err
	}

	c.incogniaToken = &result

	return nil
}
