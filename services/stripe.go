package services

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/client"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
)

var ErrNotSubscription = errors.New("no subscription found")

type stripeStore interface {
	Txer

	InsertCustomer(ctx context.Context, customer datastore.InsertCustomerParams) (datastore.Customer, error)
	GetCustomer(ctx context.Context, user *domain.User) (datastore.Customer, error)
	GetCustomerByStripeId(ctx context.Context, id string) (datastore.Customer, error)
	InsertSubscription(ctx context.Context, subscription datastore.InsertSubscriptionParams) error
	UpdateSubscription(ctx context.Context, subscription datastore.UpdateSubscriptionParams) error
	ListSubscriptions(ctx context.Context, user *domain.User, status string) ([]datastore.Subscription, error)
}

type stripeService struct {
	client *client.API
	store  stripeStore

	returnURL string
}

func NewStripe(key string, store stripeStore, domain string, tls bool) *stripeService {
	client := &client.API{}
	client.Init(key, nil)

	return &stripeService{
		client:    client,
		store:     store,
		returnURL: customerRedirectURL(domain, tls),
	}
}

func (s *stripeService) HandleSessionCheckout(ctx context.Context, session *stripe.CheckoutSession) error {
	txFunx := func(ctx context.Context) error {
		customer, err := s.store.GetCustomerByStripeId(ctx, session.Customer.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if errors.Is(err, pgx.ErrNoRows) {
			userGUID, err := uuid.Parse(session.ClientReferenceID)
			if err != nil {
				return err
			}
			customer, err = s.store.InsertCustomer(ctx, datastore.InsertCustomerParams{
				UserID: pgtype.UUID{
					Bytes: userGUID,
					Valid: true,
				},
				StripeID: session.Customer.ID,
			})
			if err != nil {
				return err
			}
		}

		params := &stripe.SubscriptionParams{}
		params.AddExpand("items.data.price.product")
		sub, err := s.client.Subscriptions.Get(session.Subscription.ID, params)
		if err != nil {
			return err
		}

		err = s.store.InsertSubscription(ctx, datastore.InsertSubscriptionParams{
			StripeID:          sub.ID,
			CustomerID:        customer.UserID,
			StripePriceID:     sub.Items.Data[0].Price.ID,
			StripeProductName: sub.Items.Data[0].Price.Product.Name,
			Status:            string(sub.Status),
			Quantity:          int32(sub.Items.Data[0].Quantity), //FIXME: truncate int??
		})

		return err
	}

	return s.store.Tx(ctx, txFunx, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
}

func (s stripeService) HandleSubscriptionUpdated(ctx context.Context, sub *stripe.Subscription) error {
	params := &stripe.SubscriptionParams{}
	params.AddExpand("items.data.price.product")
	sub, err := s.client.Subscriptions.Get(sub.ID, params)
	if err != nil {
		return err
	}

	if err := s.store.UpdateSubscription(ctx, datastore.UpdateSubscriptionParams{
		StripeID:          sub.ID,
		Status:            string(sub.Status),
		StripePriceID:     sub.Items.Data[0].Price.ID,
		StripeProductName: sub.Items.Data[0].Price.Product.Name,
		Quantity:          int32(sub.Items.Data[0].Quantity), //FIXME: truncate int??
	}); err != nil {
		return fmt.Errorf("error updating subscription: %w", err)
	}
	return nil
}

// GetSubscription returns the active subscription for the user
func (s *stripeService) GetSubscription(ctx context.Context, user *domain.User) (*domain.Subscription, error) {
	rows, err := s.store.ListSubscriptions(ctx, user, "active")
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNotSubscription
	}

	params := &stripe.SubscriptionParams{}
	params.AddExpand("items.data.price.product")
	sub, err := s.client.Subscriptions.Get(rows[0].StripeID, params)
	if err != nil {
		return nil, err
	}

	return domain.NewSubscription(sub), nil
}

func (s *stripeService) GenerateCustomerPortalURL(ctx context.Context, user *domain.User) (string, error) {

	customer, err := s.store.GetCustomer(ctx, user)
	if err != nil {
		return "", fmt.Errorf("error getting customer: %w", err)
	}

	var configurationID *string
	yes := true
	iter := s.client.BillingPortalConfigurations.List(&stripe.BillingPortalConfigurationListParams{
		ListParams: stripe.ListParams{},
		Active:     &yes,
		IsDefault:  &yes,
	})

	for iter.Next() {
		config, ok := iter.Current().(*stripe.BillingPortalConfiguration)
		if ok {
			configurationID = &config.ID
		}
	}

	sess, err := s.client.BillingPortalSessions.New(&stripe.BillingPortalSessionParams{
		Params:        stripe.Params{},
		Configuration: configurationID,
		Customer:      &customer.StripeID,
		ReturnURL:     &s.returnURL,
	})
	if err != nil {
		return "", fmt.Errorf("error creating billing portal session: %w", err)
	}

	return sess.URL, nil
}

func (s *stripeService) CreateCheckoutSession(ctx context.Context, user *domain.User, priceID string) (string, error) {
	customer, err := s.store.GetCustomer(ctx, user)
	var customerID *string
	if err == nil {
		customerID = &customer.StripeID
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("failed to get customer: %w", err)
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL:        stripe.String(s.returnURL + "?success=true"),
		CancelURL:         stripe.String(s.returnURL + "?canceled=true"),
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID: stripe.String(user.GUID.String()),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
	}

	if customerID != nil {
		params.Customer = customerID
	} else {
		params.CustomerEmail = stripe.String(user.Email)
	}

	sess, err := s.client.CheckoutSessions.New(params)
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}

func (s *stripeService) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	params := &stripe.PriceListParams{}
	params.Active = stripe.Bool(true)
	params.AddExpand("data.product")

	iter := s.client.Prices.List(params)
	var plans []domain.Plan

	for iter.Next() {
		p := iter.Price()
		if p == nil || p.Recurring == nil {
			continue
		}

		if p.Product == nil || p.Product.Deleted {
			continue
		}
		if !p.Product.Active {
			continue
		}

		plan := domain.Plan{
			ID:          p.Product.ID,
			Name:        p.Product.Name,
			Description: p.Product.Description,
			Price:       float64(p.UnitAmount) / 100.0,
			Currency:    string(p.Currency),
			Interval:    string(p.Recurring.Interval),
			PriceID:     p.ID,
			Features:    domain.NewPlanFeature(p.Product.Metadata),
		}

		plans = append(plans, plan)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error listing prices: %w", err)
	}

	slices.SortFunc(plans, func(a domain.Plan, b domain.Plan) int {
		if a.Price < b.Price {
			return -1
		} else if a.Price > b.Price {
			return 1
		} else {
			return 0
		}
	})

	return plans, nil
}

func customerRedirectURL(domain string, tls bool) string {
	url := fmt.Sprintf("%s/account", domain)
	if tls {
		return fmt.Sprintf("https://%s", url)
	} else {
		return fmt.Sprintf("http://%s", url)
	}
}
