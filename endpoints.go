package incognia

const (
	tokenEndpoint        = "/v1/token"
	signupsEndpoint      = "/v2/onboarding/signups"
	transactionsEndpoint = "/v2/authentication/transactions"
	feedbackEndpoint     = "/v2/feedbacks"
)

var (
	baseEndpoint = "https://api.incognia.com/api"
)

type endpoints struct {
	Token        string
	Signups      string
	Transactions string
	Feedback     string
}

func getEndpoints() endpoints {
	return endpoints{
		Token:        baseEndpoint + tokenEndpoint,
		Signups:      baseEndpoint + signupsEndpoint,
		Transactions: baseEndpoint + transactionsEndpoint,
		Feedback:     baseEndpoint + feedbackEndpoint,
	}
}
