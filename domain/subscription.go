package domain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v78"
)

var ErrNotSubscriptionItem = errors.New("no subscription item found")

type StripEvent struct {
	ID              string
	Type            string
	Data            json.RawMessage
	UserReferenceID string
}

type Subscription struct {
	subscription *stripe.Subscription
}

func NewSubscription(sub *stripe.Subscription) *Subscription {
	return &Subscription{
		subscription: sub,
	}
}

func (s *Subscription) Subscription() *stripe.Subscription {
	return s.subscription
}

func (s *Subscription) Price() *stripe.Price {
	return s.subscription.Items.Data[0].Price
}

func (s *Subscription) Product() *stripe.Product {
	return s.Price().Product
}

func (s *Subscription) FormatPrice(format string) string {
	return fmt.Sprintf(format, s.Price().UnitAmountDecimal/100)
}
