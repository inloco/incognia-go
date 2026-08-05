package incognia

import (
	"encoding/json"
	"testing"
)

func TestPaymentMethodACHSerializesType(t *testing.T) {
	body := postTransactionRequestBody{
		Type:      paymentType,
		AccountID: "account-id",
		PaymentMethods: []*PaymentMethod{
			{Type: ACH},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"type":"payment","account_id":"account-id","payment_methods":[{"type":"ach"}]}`
	if string(data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data))
	}
}
