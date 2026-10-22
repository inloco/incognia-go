package incognia

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These account identifiers are synthetic placeholders.
func internationalAccountFixture() *FinancialAccount {
	branch := "example_branch"
	return &FinancialAccount{
		AccountIdentifier: "example_account",
		AccountSchema:     AccountSchemaLocal,
		AccountType:       "checking",
		AccountPurpose:    "general",
		BranchIdentifier:  &branch,
		BankIdentifier:    "example_bank",
		Country:           "BR",
		HolderID:          &PersonID{Type: "example_type", Value: "example_holder"},
		HolderType:        "individual",
		PixKeys:           []*PixKey{{Type: "email", Value: "example@example.com"}},
	}
}

func TestInternationalFinancialAccountJSON(t *testing.T) {
	data, err := json.Marshal(internationalAccountFixture())
	if !assert.NoError(t, err) {
		return
	}
	assert.JSONEq(t, `{
		"account_identifier":"example_account", "account_schema":"local",
		"account_type":"checking", "account_purpose":"general",
		"branch_identifier":"example_branch", "bank_identifier":"example_bank",
		"country":"BR", "holder_id":{"type":"example_type","value":"example_holder"},
		"holder_type":"individual", "pix_keys":[{"type":"email","value":"example@example.com"}]
	}`, string(data))
}

func TestInternationalRequestJSON(t *testing.T) {
	account := &FinancialAccount{AccountIdentifier: "example_account", AccountSchema: AccountSchemaSwift}
	cases := []struct {
		name     string
		body     any
		expected string
	}{
		{"transaction", postTransactionRequestBody{
			Type: paymentType, AccountID: "example_id", FinancialAccount: account,
			CounterpartyFinancialAccount: account, CounterpartyRole: CounterpartySender,
		}, `{"type":"payment","account_id":"example_id","financial_account":{"account_identifier":"example_account","account_schema":"swift"},"counterparty_financial_account":{"account_identifier":"example_account","account_schema":"swift"},"counterparty_role":"sender"}`},
		{"feedback", postFeedbackRequestBody{
			Event: PaymentAccepted, FinancialAccount: &FeedbackFinancialAccount{FinancialAccount: *account},
			CounterpartyFinancialAccount: account, CounterpartyRole: CounterpartyRecipient,
		}, `{"event":"payment_accepted","financial_account":{"account_identifier":"example_account","account_schema":"swift"},"counterparty_financial_account":{"account_identifier":"example_account","account_schema":"swift"},"counterparty_role":"recipient"}`},
		{"transaction without accounts", postTransactionRequestBody{Type: loginType, AccountID: "example_id"}, `{"type":"login","account_id":"example_id"}`},
		{"feedback without accounts", postFeedbackRequestBody{Event: LoginAccepted}, `{"event":"login_accepted"}`},
		{"legacy feedback", postFeedbackRequestBody{
			Event: PaymentAccepted, FinancialAccount: &FeedbackFinancialAccount{
				AccountNumber: "example_number", AccountCheckDigit: "example_digit",
				BranchCode: "example_branch", IspbCode: "example_bank",
				HolderTaxID: &PersonID{Type: "example_type", Value: "example_holder"},
			},
		}, `{"event":"payment_accepted","financial_account":{"account_number":"example_number","account_check_digit":"example_digit","branch_code":"example_branch","ispb_code":"example_bank","holder_tax_id":{"type":"example_type","value":"example_holder"}}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.body)
			if !assert.NoError(t, err) {
				return
			}
			assert.JSONEq(t, tc.expected, string(data))
		})
	}
}

func (suite *IncogniaTestSuite) TestInternationalTransactionAccounts() {
	account := internationalAccountFixture()
	counterparty := &FinancialAccount{AccountIdentifier: "example_counterparty", AccountSchema: AccountSchemaSwift, Country: "GB"}
	for _, role := range []CounterpartyRole{CounterpartySender, CounterpartyRecipient} {
		for _, withContext := range []bool{false, true} {
			for _, kind := range []string{"payment", "login", "web login"} {
				expected := &postTransactionRequestBody{
					Type: loginType, AccountID: "example_id", FinancialAccount: account,
					CounterpartyFinancialAccount: counterparty, CounterpartyRole: role,
				}
				if kind == "payment" {
					expected.Type = paymentType
				}
				server := suite.mockPostTransactionsEndpoint(token, expected, transactionAssessmentFixture, emptyQueryString)
				var err error
				switch kind {
				case "payment":
					params := &Payment{AccountID: "example_id", FinancialAccount: account, CounterpartyFinancialAccount: counterparty, CounterpartyRole: role}
					if withContext {
						_, err = suite.client.RegisterPaymentWithContext(context.Background(), params)
					} else {
						_, err = suite.client.RegisterPayment(params)
					}
				case "login":
					params := &Login{AccountID: "example_id", FinancialAccount: account, CounterpartyFinancialAccount: counterparty, CounterpartyRole: role}
					if withContext {
						_, err = suite.client.RegisterLoginWithContext(context.Background(), params)
					} else {
						_, err = suite.client.RegisterLogin(params)
					}
				case "web login":
					params := &WebLogin{AccountID: "example_id", FinancialAccount: account, CounterpartyFinancialAccount: counterparty, CounterpartyRole: role}
					if withContext {
						_, err = suite.client.RegisterWebLoginWithContext(context.Background(), params)
					} else {
						_, err = suite.client.RegisterWebLogin(params)
					}
				}
				server.Close()
				suite.NoError(err, "%s, role=%s, context=%t", kind, role, withContext)
			}
		}
	}
}

func (suite *IncogniaTestSuite) TestInternationalAndLegacyFeedbackAccounts() {
	modern := &FeedbackFinancialAccount{FinancialAccount: *internationalAccountFixture()}
	legacy := &FeedbackFinancialAccount{AccountNumber: "example_number", BranchCode: "example_branch", IspbCode: "example_bank"}
	oldAccount := &BankAccountInfo{AccountNumber: "example_number"}
	for _, account := range []*FeedbackFinancialAccount{modern, legacy} {
		for _, role := range []CounterpartyRole{CounterpartySender, CounterpartyRecipient} {
			expected := &postFeedbackRequestBody{
				Event: PaymentAccepted, FinancialAccount: account, CounterpartyRole: role,
				CounterpartyFinancialAccount: internationalAccountFixture(), DebtorAccount: oldAccount, CreditorAccount: oldAccount,
			}
			identifiers := &FeedbackIdentifiers{
				FinancialAccount: account, CounterpartyRole: role,
				CounterpartyFinancialAccount: expected.CounterpartyFinancialAccount, DebtorAccount: oldAccount, CreditorAccount: oldAccount,
			}
			calls := []func() error{
				func() error { return suite.client.RegisterFeedback(PaymentAccepted, nil, identifiers) },
				func() error {
					return suite.client.RegisterFeedbackWithContext(context.Background(), PaymentAccepted, nil, identifiers)
				},
				func() error {
					return suite.client.RegisterFeedbackWithExpiration(PaymentAccepted, nil, nil, identifiers)
				},
				func() error {
					return suite.client.RegisterFeedbackWithExpirationWithContext(context.Background(), PaymentAccepted, nil, nil, identifiers)
				},
			}
			for _, call := range calls {
				server := suite.mockFeedbackEndpoint(token, expected)
				err := call()
				server.Close()
				suite.NoError(err)
			}
		}
	}
}
