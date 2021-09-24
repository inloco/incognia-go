package incognia

const (
	tokenEndpoint        = "/v1/token"
	signupsEndpoint      = "/v2/onboarding/signups"
	transactionsEndpoint = "/v2/authentication/transactions"
)

type endpoints struct {
	Token        string
	Signups      string
	Transactions string
}

var baseEndpoint string = "https://api.us.incognia.com/api"

func buildEndpoints() endpoints {
	return endpoints{
		Token:        baseEndpoint + tokenEndpoint,
		Signups:      baseEndpoint + signupsEndpoint,
		Transactions: baseEndpoint + transactionsEndpoint,
	}
}
