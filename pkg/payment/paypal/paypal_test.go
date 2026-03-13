package paypal_test

import (
	"rest_demo/pkg/payment/paypal"
	"testing"
)

func getPaypal(t *testing.T) *paypal.Paypal {
	p, err := paypal.NewPaypal(paypal.Config{
		Endpoint: "https://api.sandbox.paypal.com",
		ClientId: "AU8ufr2ReJfToqxs2YUIXrBqrPiHEXngxVZ9Zs5QsXnahyLtVA2TKQtdmd7een0tTZ03FeHAow69KVJW",
		Secret:   "EIlslIPb83BaL4Q-ZsbbaxQanDGxuEaT-gvivgVqmkdwnb_naRT7XosHchb4I2rqPJUzDDRliYNh8XXB",
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}
