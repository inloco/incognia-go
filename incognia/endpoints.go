package incognia

var baseEndpoint string = "https://api.us.incognia.com/api"

var (
	tokenEndpoint        = baseEndpoint + "/v1/token"
	signupsEndpoint      = baseEndpoint + "/v2/onboarding/signups"
	transactionsEndpoint = baseEndpoint + "/v2/authentication/transactions"
)
