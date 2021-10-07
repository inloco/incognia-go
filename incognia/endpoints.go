package incognia

const (
	tokenEndpoint        = "/v1/token"
	signupsEndpoint      = "/v2/onboarding/signups"
	transactionsEndpoint = "/v2/authentication/transactions"
	feedbackEndpoint     = "/v2/feedbacks"
)

var baseEndpoint map[Region]string = map[Region]string{
	US: "https://api.us.incognia.com/api",
	BR: "https://api.br.incognia.com/api",
}

type endpoints struct {
	Token        string
	Signups      string
	Transactions string
	Feedback     string
}

func newEndpoints(region Region) endpoints {
	return endpoints{
		Token:        baseEndpoint[region] + tokenEndpoint,
		Signups:      baseEndpoint[region] + signupsEndpoint,
		Transactions: baseEndpoint[region] + transactionsEndpoint,
		Feedback:     baseEndpoint[region] + feedbackEndpoint,
	}
}
