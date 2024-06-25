package services

import (
	"context"
	"errors"

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
	GetCustomer(ctx context.Context, id string) (datastore.Customer, error)
	InsertSubscription(ctx context.Context, subscription datastore.InsertSubscriptionParams) error
	ListSubscriptions(ctx context.Context, user *domain.User, status string) ([]datastore.Subscription, error)
}

type stripeService struct {
	client *client.API
	store  stripeStore
}

func NewStripeService(key string, store stripeStore) *stripeService {
	client := &client.API{}
	client.Init(key, nil)

	return &stripeService{
		client: client,
		store:  store,
	}
}
func (s *stripeService) HandlerSessionCheckout(ctx context.Context, session *stripe.CheckoutSession) error {
	txFunx := func(ctx context.Context) error {
		customer, err := s.store.GetCustomer(ctx, session.Customer.ID)
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

// GetSubscription returns the active subscription for the user
// func (s *stripeService) GetSubscription(ctx context.Context, user *domain.User) (*domain.Subscription, error) {
// 	customerID := user.ReferenceID()
// 	events, err := s.store.ListStripeEvent(ctx, customerID, "checkout.session.completed")
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(events) == 0 {
// 		return nil, ErrNotSubscription
// 	}

// 	sort.Slice(events, func(i, j int) bool {
// 		return events[i].CreatedAt.Time.After(events[j].CreatedAt.Time)
// 	})

// 	wrapper := struct {
// 		Data struct {
// 			Object struct {
// 				ID           string                    `json:"id"`
// 				Status       stripe.SubscriptionStatus `json:"status"`
// 				Subscription string                    `json:"subscription"`
// 			} `json:"object"`
// 		} `json:"data"`
// 	}{}
// 	for _, e := range events {
// 		fmt.Println(string(e.Data))
// 		if err := json.Unmarshal(e.Data, &wrapper); err != nil {
// 			return nil, fmt.Errorf("error unmarshalling session: %w", err)
// 		}

// 		params:=stripe.SubscriptionParams{}
// 		params.AddExpand("

//TODO: handler all possible states
// if wrapper.Data.Object.Status == stripe.SubscriptionStatusActive {
// 	return domain.LoadSubscription(wrapper.Data.Object.Subscription, s.client)
// }
// 	}

// 	return nil, ErrNotSubscription
// }
