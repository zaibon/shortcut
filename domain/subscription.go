package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
	Features     PlanFeature
}

func NewSubscription(sub *stripe.Subscription) *Subscription {
	return &Subscription{
		subscription: sub,
		Features:     NewPlanFeature(sub.Items.Data[0].Price.Product.Metadata),
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

type Plan struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Currency    string
	Interval    string
	PriceID     string
	Features    PlanFeature
}

func NewPlanFeature(metadata map[string]string) PlanFeature {
	pf := PlanFeature{}

	promoted, ok := metadata["promoted"]
	if ok && promoted == "true" {
		pf.Promoted = true
	}

	linksNumber, ok := metadata["links"]
	if ok {
		if number, err := strconv.Atoi(linksNumber); err == nil {
			pf.LinksNumber = number
		}
	}

	return pf
}

type PlanFeature struct {
	Promoted    bool
	LinksNumber int
}
