# Incognia API Go Client

![test workflow](https://github.com/inloco/incognia-go/actions/workflows/continuous.yml/badge.svg)

Go lightweight client library for [Incognia APIs](https://dash.incognia.com/api-reference).

## Installation

```
go get repo.incognia.com/go/incognia
```

## Usage

### Configuration

First, you need to obtain an instance of the API client using `New`. It receives a configuration
object of `IncogniaClientConfig` that contains the following parameters:

| Parameter             | Description                                    | Required | Default       |
| --------------------- | ---------------------------------------------- | -------- | ------------- |
| `ClientID`            | Your client ID                                 | **Yes**  | -             |
| `ClientSecret`        | Your client secret                             | **Yes**  | -             |
| `Timeout`             | Request timeout                                | **No**   | 10 seconds    |
| `HTTPClient`          | Custom HTTP client                             | **No**   | `http.Client` |

For instance, if you need the default client:

```go
client, err := incognia.New(&incognia.IncogniaClientConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
})
if err != nil {
    log.Fatal("could not initialize Incognia client")
}
```

or if you need a client that uses a specific timeout:

```go
client, err := incognia.New(&incognia.IncogniaClientConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Timeout:      time.Second * 2,
})
if err != nil {
    log.Fatal("could not initialize Incognia client")
}
```

or if you need a custom HTTP client:

```go
transport := http.DefaultTransport.(*http.Transport).Clone()
transport.MaxIdleConns = 1000
transport.MaxIdleConnsPerHost = 100
transport.MaxConnsPerHost = 200

httpClient := &http.Client{
    Timeout:   time.Second * 2,
    Transport: transport,
}

client, err := incognia.New(&incognia.IncogniaClientConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Timeout:      time.Second * 2,
    HTTPClient:   httpClient,
})
if err != nil {
    log.Fatal("could not initialize Incognia client")
}
```

### Incognia API

The implementation is based on the [Incognia API Reference](https://dash.incognia.com/api-reference).

### Authentication

Authentication is done transparently, so you don't need to worry about it.

### Registering Signup

This method registers a new signup for the given installation and address, returning a `SignupAssessment`, containing the risk assessment and supporting evidence:

```go
assessment, err := client.RegisterSignup("installation-id", &incognia.Address{
    AddressLine: "20 W 34th St, New York, NY 10001, United States",
    StructuredAddress: &incognia.StructuredAddress{
        Locale:       "en-US",
        CountryName:  "United States of America",
        CountryCode:  "US",
        State:        "NY",
        City:         "New York City",
        Borough:      "Manhattan",
        Neighborhood: "Midtown",
        Street:       "W 34th St.",
        Number:       "20",
        Complements:  "Floor 2",
        PostalCode:   "10001",
    },
    Coordinates: &incognia.Coordinates{
        Lat: -23.561414,
        Lng: -46.6558819,
    },
})
```

To provide additional parameters like policy id (optional) and account id (optional), use the `RegisterSignupWithParams` method:

```go
assessment, err := client.RegisterSignupWithParams(&incognia.Signup{
	InstallationID: "installation-id",//required
	Address: &incognia.Address{//optional, use nil if you don't have an address
        AddressLine: "20 W 34th St, New York, NY 10001, United States",
        StructuredAddress: &incognia.StructuredAddress{
            Locale:       "en-US",
            CountryName:  "United States of America",
            CountryCode:  "US",
            State:        "NY",
            City:         "New York City",
            Borough:      "Manhattan",
            Neighborhood: "Midtown",
            Street:       "W 34th St.",
            Number:       "20",
            Complements:  "Floor 2",
            PostalCode:   "10001",
        },
        Coordinates: &incognia.Coordinates{
            Lat: -23.561414,
            Lng: -46.6558819,
        },
    },
    AccountID: "account-id",//optional, use empty string if you don't have an account id
    PolicyID:  "policy-id",//optional, use empty string if you don't have a policy id
})
```

### Registering Payment

This method registers a new payment for the given installation and account, returning a `TransactionAssessment`, containing the risk assessment and supporting evidence.

```go
assessment, err := client.RegisterPayment(&incognia.Payment{
    InstallationID: "installation-id",
    AccountID:      "account-id",
    ExternalID:     "external-id",
    PolicyID:       "policy-id",
    Addresses: []*incognia.TransactionAddress{
        {
            Type: incognia.Billing,
            AddressLine:    "20 W 34th St, New York, NY 10001, United States",
            StructuredAddress: &incognia.StructuredAddress{
                Locale:       "en-US",
                CountryName:  "United States of America",
                CountryCode:  "US",
                State:        "NY",
                City:         "New York City",
                Borough:      "Manhattan",
                Neighborhood: "Midtown",
                Street:       "W 34th St.",
                Number:       "20",
                Complements:  "Floor 2",
                PostalCode:   "10001",
            },
            Coordinates: &incognia.Coordinates{
                Lat: -23.561414,
                Lng: -46.6558819,
            },
        },
    },
    Value: &incognia.PaymentValue{
        Amount:   55.02,
        Currency: "BRL",
    },
    Methods: []*incognia.PaymentMethod{
        {
	    Type: incognia.GooglePay,
	},
        {
            Type: incognia.CreditCard,
            CreditCard: &incognia.CardInfo{
                Bin:            "292821",
                LastFourDigits: "2222",
                ExpiryYear:     "2020",
                ExpiryMonth:    "10",
            },
        },
    },
})
```

This method registers a new **web** payment for the given installation and account, returning a `TransactionAssessment`, containing the risk assessment and supporting evidence.

```go
assessment, err := client.RegisterPayment(&incognia.Payment{
    RequestToken:   "request-token",
    AccountID:      "account-id",
    ExternalID:     "external-id",
    PolicyID:       "policy-id",
    ...
})
```


### Registering Login

This method registers a new login for the given installation and account, returning a `TransactionAssessment`, containing the risk assessment and supporting evidence.

```go
assessment, err := client.RegisterLogin(&incognia.Login{
    InstallationID:             "installation-id",
    AccountID:                  "account-id",
    ExternalID:                 "external-id",
    PolicyID:                   "policy-id",
    PaymentMethodIdentifier:    "payment-method-identifier",
})
```

This method registers a new **web** login for the given account and request-token, returning a `TransactionAssessment`, containing the risk assessment and supporting evidence.

```go
assessment, err := client.RegisterLogin(&incognia.Login{
    RequestToken:               "request-token",
    AccountID:                  "account-id",
    ...
})
```

### Registering Payment or Login without evaluating its risk assessment

Turning off the risk assessment evaluation allows you to register a new transaction (Login or Payment), but the response (`TransactionAssessment`) will be empty. For instance, if you're using the risk assessment only for some payment transactions, you should still register all the other ones: this will avoid any bias on the risk assessment computation.

To register a login or a payment without evaluating its risk assessment, you should use the `Eval *bool` attribute as follows:

Login example:

```go
shouldEval := false

assessment, err := client.RegisterLogin(&incognia.Login{
    Eval:           &shouldEval,
    InstallationID: "installation-id",
    AccountID:      "account-id",
    ExternalID:     "external-id",
    PolicyID:       "policy-id",
})
```

Payment example:

```go
shouldEval := false

assessment, err := client.RegisterPayment(&incognia.Payment{
    Eval:            &shouldEval,
    InstallationID: "installation-id",
    AccountID:      "account-id",
    ExternalID:     "external-id",
    PolicyID:       "policy-id",
    Addresses: []*incognia.TransactionAddress{
        {
            Type: incognia.Billing,
            AddressLine:    "20 W 34th St, New York, NY 10001, United States",
            StructuredAddress: &incognia.StructuredAddress{
                Locale:       "en-US",
                CountryName:  "United States of America",
    ...
```

### Sending Feedback

This method registers a feedback event for the given identifiers (represented in `FeedbackIdentifiers`) related to a signup, login or payment.

```go
occurredAt, err := time.Parse(time.RFC3339, "2024-07-22T15:20:00Z")
feedbackEvent := incognia.AccountTakeover
err := client.RegisterFeedback(feedbackEvent, &occurredAt, &incognia.FeedbackIdentifiers{
    InstallationID: "some-installation-id",
    AccountID:      "some-account-id",
})
```

### Authentication

Our library authenticates clients automatically, but clients may want to authenticate manually because our token route has a long response time (to avoid brute force attacks). If that's your case, you can choose the moment which authentication occurs by leveraging `ManualRefreshTokenProvider`, as shown by the example:

```go
tokenClient := incognia.NewTokenClient(&TokenClientConfig{clientID: clientID, clientSecret: clientSecret})
tokenProvider := incognia.NewManualRefreshTokenProvider(tokenClient)
c, err := incognia.New(&IncogniaClientConfig{TokenProvider: tokenProvider})
if err != nil {
    log.Fatal("could not initialize Incognia client")
}

go func(i *incognia.Client) {
  for {
      accessToken, err := tokenProvider.Refresh()
      if (err != nil) {
          log.PrintLn("could not refresh incognia token")
          continue
      }
      time.Sleep(time.Until(accessToken.GetExpiresAt()))
   }
}(c)
```

You can also keep the default automatic authentication but increase the token route timeout by changing the `TokenRouteTimeout` parameter of your `IncogniaClientConfig`.

## Evidences

Every assessment response (`TransactionAssessment` and `SignupAssessment`) includes supporting evidence in the type `Evidence`, which provides methods `GetEvidence` and `GetEvidenceAsInt64` to help you getting and parsing values. You can see usage examples below:

```go
var deviceModel string

err := assessment.Evidence.GetEvidence("device_model", &deviceModel)
if err != nil {
    return err
}

fmt.Println(deviceModel)
```

You can also access specific evidences using their full path. For example, to get `location_permission_enabled` evidence from the following response:

```json
{
    "evidence": {
        "location_services": {
            "location_permission_enabled": true
        }
    }
}
```

call any type of `GetEvidence` method using the evidence's full path:

```go
var locationPermissionEnabled bool
err := assessment.Evidence.GetEvidence("location_services.location_permission_enabled", &locationPermissionEnabled)
if err != nil {
    return err
}

fmt.Println(locationPermissionEnabled)
```

You can find all available evidence [here](https://developer.incognia.com/docs/apis/v2/understanding-assessment-evidence).

## How to Contribute

If you have found a bug or if you have a feature request, please report them at this repository issues section.

## What is Incognia?

Incognia is a location identity platform for mobile apps that enables:

- Real-time address verification for onboarding
- Frictionless authentication
- Real-time transaction verification

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
