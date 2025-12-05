package handlers

import (
	"context"
	"net/http"

	"github.com/donseba/go-htmx"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v78"

	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/middleware"
	"github.com/zaibon/shortcut/templates"
)

type stripeService interface {
	GetSubscription(ctx context.Context, user *domain.User) (*domain.Subscription, error)
	HandleSessionCheckout(ctx context.Context, session *stripe.CheckoutSession) error
	HandleSubscriptionUpdated(ctx context.Context, sub *stripe.Subscription) error
	GenerateCustomerPortalURL(ctx context.Context, user *domain.User) (string, error)
}

type subscriptionHandlers struct {
	htmx    *htmx.HTMX
	payment stripeService
	urls    URLService

	stripeKey            string
	stripeEndpointSecret string
}

func NewSubscriptionHandlers(
	stripeKey,
	stripeEndpointSecret string,
	payment stripeService,
	urls URLService,
) *subscriptionHandlers {
	return &subscriptionHandlers{
		htmx:                 htmx.New(),
		stripeKey:            stripeKey,
		stripeEndpointSecret: stripeEndpointSecret,
		payment:              payment,
		urls:                 urls,
	}
}

func (h *subscriptionHandlers) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated)
	})
	// r.Post("/subscription/webhook", h.webhook)
	r.Get("/subscription", h.subscription)
}

func (h *subscriptionHandlers) subscription(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	urlCount, err := h.urls.CountMonthlyURL(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	limit := domain.FreePlanLimit
	stats := domain.SubscriptionStats{
		PlanName:        "Free",
		URLUsage:        int(urlCount),
		URLLimit:        limit,
		Remaining:       limit - int(urlCount),
		UsagePercentage: int((float64(urlCount) / float64(limit)) * 100),
	}
	if stats.UsagePercentage > 100 {
		stats.UsagePercentage = 100
	}
	if stats.Remaining < 0 {
		stats.Remaining = 0
	}

	templates.SubscriptionPage(stats).
		Render(r.Context(), w)
}

// func (h *subscriptionHandlers) webhook(w http.ResponseWriter, r *http.Request) {
// 	const MaxBodyBytes = int64(65536)
// 	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
// 	payload, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		log.Error("Error reading request body", "err", err)
// 		w.WriteHeader(http.StatusServiceUnavailable)
// 		return
// 	}

// 	event, err := h.verifyWebhookSignature(payload, r.Header.Get("Stripe-Signature"))
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
// 		return
// 	}

// 	log.Info("Received event", "type", event.Type)

// 	// Unmarshal the event data into an appropriate struct depending on its Type
// 	switch event.Type {
// 	case "checkout.session.completed":
// 		if err := h.handlerSessionCompleted(r.Context(), event); err != nil {
// 			log.Error("Error handling session completed", "err", err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 	case "customer.subscription.updated", "customer.subscription.deleted":
// 		if err := h.handleSubscriptionUpdated(r.Context(), event); err != nil {
// 			log.Error("Error handling subscription updated", "err", err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 	default:
// 		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
// 	}

// 	w.WriteHeader(http.StatusOK)
// }

// func (h *subscriptionHandlers) verifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
// 	// Pass the request body and Stripe-Signature header to ConstructEvent, along
// 	// with the webhook signing key.
// 	event, err := webhook.ConstructEvent(payload, signature, h.stripeEndpointSecret)

// 	if err != nil {
// 		log.Error("Error verifying webhook signature", "err", err)
// 		return nil, err
// 	}
// 	return &event, nil
// }

// func (h *subscriptionHandlers) handlerSessionCompleted(ctx context.Context, event *stripe.Event) error {
// 	var session stripe.CheckoutSession
// 	err := json.Unmarshal(event.Data.Raw, &session)
// 	if err != nil {
// 		return fmt.Errorf("error parsing webhook JSON: %w", err)
// 	}

// 	if err := h.svc.HandleSessionCheckout(ctx, &session); err != nil {
// 		return fmt.Errorf("error handling checkout session completed event: %w", err)
// 	}

// 	return nil
// }

// func (h *subscriptionHandlers) handleSubscriptionUpdated(ctx context.Context, event *stripe.Event) error {
// 	var sub stripe.Subscription
// 	err := json.Unmarshal(event.Data.Raw, &sub)
// 	if err != nil {
// 		return fmt.Errorf("error parsing webhook JSON: %w", err)
// 	}

// 	if err := h.svc.HandleSubscriptionUpdated(ctx, &sub); err != nil {
// 		return fmt.Errorf("error handling checkout session completed event: %w", err)
// 	}

// 	return nil
// }
