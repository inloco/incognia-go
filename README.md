# Incognia API Go Client
![test workflow](https://github.com/inloco/incognia-go/actions/workflows/continuous.yml/badge.svg)

Go lightweight client library for [Incognia APIs](https://dash.incognia.com/api-reference).

## Installation

```
go get repo.incognia.com/go/incognia
```

## Usage

### Configuration

Before calling the API methods, you need to create an instance of the `Client` struct.

```go
// to use the US region
client, err := incognia.New(&incognia.IncogniaClientConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Region:       incognia.US,
})

// to use the BR region
client, err := incognia.New(&incognia.IncogniaClientConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Region:       incognia.BR,
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

### Getting a Signup

This method allows you to query the latest assessment for a given signup event, returning a `SignupAssessment`, containing the risk assessment and supporting evidence:

```go
signupID := "c9ac2803-c868-4b7a-8323-8a6b96298ebe"
assessment, err := client.GetSignupAssessment(signupID)
```

### Registering Payment

This method registers a new payment for the given installation and account, returning a `TransactionAssessment`, containing the risk assessment and supporting evidence.

```go
assessment, err := client.RegisterPayment(&incognia.Payment{
    InstallationID: "installation-id",
    AccountID:      "account-id",
    ExternalID:     "external-id",
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
            Type: incognia.CreditCard,
            CreditCard: &incognia.CardInfo{
                Bin:            "29282",
                LastFourDigits: "2222",
                ExpiryYear:     "2020",
                ExpiryMonth:    "10",
            },
        },
    },
})
```

### Registering Login

This method registers a new login for the given installation and account, returning a `TransactionAssessment`, containing the risk assessment and supporting evidence.

```go
assessment, err := client.RegisterLogin(&incognia.Login{
    InstallationID: "installation-id",
    AccountID:      "account-id",
    ExternalID:     "external-id",
})
```

### Sending Feedback

This method registers a feedback event for the given identifiers (represented in `FeedbackIdentifiers`) related to a signup, login or payment.

```go
timestamp := time.Now()
feedbackEvent := incognia.SignupAccepted
err := client.RegisterFeedback(feedbackEvent, &timestamp, &incognia.FeedbackIdentifiers{
		InstallationID: "some-installation-id",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
})
```

## Evidences

Every assessment response (`TransactionAssessment` and `SignupAssessment`) includes supporting evidence in a generic `map[string]interface{}`.
You can find all available evidence [here](https://docs.incognia.com/apis/understanding-assessment-evidence#risk-assessment-evidence).

## How to Contribute

If you have found a bug or if you have a feature request, please report them at this repository issues section.

## What is Incognia?

Incognia is a location identity platform for mobile apps that enables:

- Real-time address verification for onboarding
- Frictionless authentication
- Real-time transaction verification

## Create a Free Incognia Account

1. Go to [Incognia](https://www.incognia.com/) and click on "Get Started"
2. Fill the contact form
3. Once we contact you, you will be ready to integrate [Incognia SDK](https://docs.incognia.com/sdk/getting-started) and use [Incognia APIs](https://dash.incognia.com/api-reference)

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)