package incognia

const (
	tokenEndpoint        = "/v1/token"
	signupsEndpoint      = "/v2/onboarding/signups"
	transactionsEndpoint = "/v2/authentication/transactions"
)

type Region int64

const (
	US Region = iota
	BR
)

type endpoints struct {
	Token        string
	Signups      string
	Transactions string
}

var baseEndpoint map[Region]string = map[Region]string{
	US: "https://api.us.incognia.com/api",
	BR: "https://incognia.inloco.com.br/api",
}

func buildEndpoints(region Region) endpoints {
	return endpoints{
		Token:        baseEndpoint[region] + tokenEndpoint,
		Signups:      baseEndpoint[region] + signupsEndpoint,
		Transactions: baseEndpoint[region] + transactionsEndpoint,
	}
}
