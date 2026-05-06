package payment

import (
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
)

type StripeClient struct {
	secretKey string
}

func NewStripeClient(secretKey string) *StripeClient {
	stripe.Key = secretKey
	return &StripeClient{
		secretKey: secretKey,
	}
}

func (c *StripeClient) CreatePaymentIntent(amount int64, currency string, orderID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"order_id": orderID,
		},
	}
	return paymentintent.New(params)
}

func (c *StripeClient) GetPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(id, nil)
}
