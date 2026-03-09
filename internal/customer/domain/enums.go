package domain

// LifecycleStage represents the customer's journey stage.
type LifecycleStage string

const (
	LifecycleNewUser         LifecycleStage = "new_user"
	LifecycleInstalledNoTrip LifecycleStage = "installed_no_trip"
	LifecycleActivated       LifecycleStage = "activated" // 1 trip
	LifecycleReturning       LifecycleStage = "returning" // 2+ trips
	LifecycleLoyal           LifecycleStage = "loyal"     // 3+ trips
	LifecycleVIP             LifecycleStage = "vip"       // 5+ trips
	LifecycleDormant         LifecycleStage = "dormant"   // 30+ days inactive
)

// CustomerTier represents the customer's tier level.
type CustomerTier string

const (
	TierStandard CustomerTier = "standard"
	TierLuxury   CustomerTier = "luxury" // ≥3 trips AND total spent ≥ 1,000,000 VND
)

// LuxuryMinTrips is the minimum number of completed trips for Luxury tier.
const LuxuryMinTrips = 3

// LuxuryMinSpent is the minimum total spending (VND) for Luxury tier.
const LuxuryMinSpent float64 = 1_000_000
