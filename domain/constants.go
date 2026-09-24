package domain

const (
	TimeFormat    = "Mon, 02 Jan 06 at 15:04:05"
	FreePlanLimit = 10
	// UnlimitedPlanLimit is the effective cap for a paid plan that has no
	// explicit `links` metadata on its Stripe product.
	UnlimitedPlanLimit = 1000000
	SessionLifetime    = 2 * 24 * 60 * 60 // 2 days in seconds
)
