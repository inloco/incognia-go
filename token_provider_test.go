package incognia

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	accessTokenFixture = &accessToken{
		CreatedAt:   time.Now().Unix(),
		ExpiresIn:   int64(1600),
		AccessToken: "some-token",
		TokenType:   "Bearer",
	}
	expiredTokenFixture = &accessToken{
		CreatedAt:   time.Now().Unix(),
		ExpiresIn:   0,
		AccessToken: "some-token",
		TokenType:   "Bearer",
	}
)

type ManualRefreshTokenProviderTestSuite struct {
	suite.Suite

	tokenProvider  *ManualRefreshTokenProvider
	tokenNetClient *http.Client
}

func (suite *ManualRefreshTokenProviderTestSuite) SetupTest() {
	tokenClient := NewTokenClient(&TokenClientConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	suite.tokenProvider = NewManualRefreshTokenProvider(tokenClient)
	suite.tokenNetClient = suite.tokenProvider.tokenClient.netClient
}

func (suite *ManualRefreshTokenProviderTestSuite) TestSuccessGetToken() {
	expectedToken := accessTokenFixture
	suite.tokenProvider.token = expectedToken

	actualToken, err := suite.tokenProvider.GetToken()
	suite.NoError(err)
	suite.Equal(expectedToken, actualToken)
}

func (suite *ManualRefreshTokenProviderTestSuite) TestGetTokenNotFound() {
	_, err := suite.tokenProvider.GetToken()
	suite.EqualError(err, ErrTokenNotFound.Error())
}

func (suite *ManualRefreshTokenProviderTestSuite) TestGetTokenExpiredToken() {
	expectedToken := expiredTokenFixture
	suite.tokenProvider.token = expectedToken

	_, err := suite.tokenProvider.GetToken()
	suite.EqualError(err, ErrTokenExpired.Error())
}

func (suite *ManualRefreshTokenProviderTestSuite) TestRefreshSuccess() {
	suite.tokenProvider.tokenClient.tokenEndpoint = mockTokenEndpoint(accessTokenFixture.AccessToken, "1000").URL
	token, err := suite.tokenProvider.Refresh()
	accessToken := token.(*accessToken)
	suite.NoError(err)
	suite.Equal(accessToken.AccessToken, accessTokenFixture.AccessToken)
}

func (suite *ManualRefreshTokenProviderTestSuite) TestRefreshUnexpectedError() {
	suite.tokenProvider.tokenClient.tokenEndpoint = mockStatusServer(http.StatusInternalServerError).URL
	_, err := suite.tokenProvider.Refresh()
	suite.Error(err)
	suite.Contains(err.Error(), strconv.Itoa(http.StatusInternalServerError))
}

func (suite *ManualRefreshTokenProviderTestSuite) TestRefreshUnauthorized() {
	suite.tokenProvider.tokenClient.tokenEndpoint = mockStatusServer(http.StatusUnauthorized).URL
	_, err := suite.tokenProvider.Refresh()
	suite.EqualError(err, ErrInvalidCredentials.Error())
}

func TestManualRefreshTokenProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ManualRefreshTokenProviderTestSuite))
}

type AutoRefreshTokenProviderTestSuite struct {
	suite.Suite

	tokenProvider  *AutoRefreshTokenProvider
	tokenNetClient *http.Client
}

func (suite *AutoRefreshTokenProviderTestSuite) SetupTest() {
	tokenClient := NewTokenClient(&TokenClientConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	suite.tokenProvider = NewAutoRefreshTokenProvider(tokenClient)
	suite.tokenNetClient = suite.tokenProvider.tokenClient.netClient
}

func (suite *AutoRefreshTokenProviderTestSuite) TestSuccessGetToken() {
	expectedToken := accessTokenFixture
	suite.tokenProvider.token = expectedToken

	actualToken, err := suite.tokenProvider.GetToken()
	suite.NoError(err)
	suite.Equal(expectedToken, actualToken)
}

func (suite *AutoRefreshTokenProviderTestSuite) TestGetTokenNotFound() {
	suite.tokenProvider.tokenClient.tokenEndpoint = mockTokenEndpoint(accessTokenFixture.AccessToken, "1000").URL

	token, err := suite.tokenProvider.GetToken()
	accessToken := token.(*accessToken)
	suite.NoError(err)
	suite.Equal(accessToken.AccessToken, accessTokenFixture.AccessToken)
}

func (suite *AutoRefreshTokenProviderTestSuite) TestGetTokenExpiredToken() {
	suite.tokenProvider.token = expiredTokenFixture

	suite.tokenProvider.tokenClient.tokenEndpoint = mockTokenEndpoint(accessTokenFixture.AccessToken, "1000").URL

	token, err := suite.tokenProvider.GetToken()
	accessToken := token.(*accessToken)
	suite.NoError(err)
	suite.Equal(accessToken.AccessToken, accessTokenFixture.AccessToken)
}

func (suite *AutoRefreshTokenProviderTestSuite) TestRefreshSuccess() {
	suite.tokenProvider.tokenClient.tokenEndpoint = mockTokenEndpoint(accessTokenFixture.AccessToken, "1000").URL
	token, err := suite.tokenProvider.GetToken()
	accessToken := token.(*accessToken)
	suite.NoError(err)
	suite.Equal(accessToken.AccessToken, accessTokenFixture.AccessToken)
}

func (suite *AutoRefreshTokenProviderTestSuite) TestGetTokenUnexpectedError() {
	suite.tokenProvider.tokenClient.tokenEndpoint = mockStatusServer(http.StatusInternalServerError).URL
	_, err := suite.tokenProvider.GetToken()
	suite.Error(err)
	suite.Contains(err.Error(), strconv.Itoa(http.StatusInternalServerError))
}

func (suite *AutoRefreshTokenProviderTestSuite) TestRefreshUnauthorized() {
	suite.tokenProvider.tokenClient.tokenEndpoint = mockStatusServer(http.StatusUnauthorized).URL
	_, err := suite.tokenProvider.GetToken()
	suite.EqualError(err, ErrInvalidCredentials.Error())
}

func TestAutoRefreshTokenProviderTestSuite(t *testing.T) {
	suite.Run(t, new(AutoRefreshTokenProviderTestSuite))
}
