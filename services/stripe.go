package services

import (
	"context"
	"errors"
	"fmt"

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

func customerRedirectURL(domain string, tls bool) string {
	url := fmt.Sprintf("%s/my-account", domain)
	if tls {
		return fmt.Sprintf("https://%s", url)
	} else {
		return fmt.Sprintf("http://%s", url)
	}
}
