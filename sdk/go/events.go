package crmsdk

// Event is the interface all trackable events must implement.
type Event interface {
	EventType() string
	UserID() string
}

// --- User Lifecycle Events ---

// EventUserRegistered is sent when a new user registers.
type EventUserRegistered struct {
	ExternalUserID string `json:"external_user_id"`
	Phone          string `json:"phone"`
	FullName       string `json:"full_name"`
	Email          string `json:"email,omitempty"`
	Source         string `json:"source,omitempty"`
	CampaignID     string `json:"campaign_id,omitempty"`
	DeviceType     string `json:"device_type,omitempty"`
}

func (e EventUserRegistered) EventType() string { return "user_registered" }
func (e EventUserRegistered) UserID() string    { return e.ExternalUserID }

// EventAppInstalled is sent when the app is installed.
type EventAppInstalled struct {
	ExternalUserID string `json:"external_user_id"`
	DeviceType     string `json:"device_type,omitempty"`
}

func (e EventAppInstalled) EventType() string { return "app_installed" }
func (e EventAppInstalled) UserID() string    { return e.ExternalUserID }

// EventAppOpened is sent when the app is opened.
type EventAppOpened struct {
	ExternalUserID string `json:"external_user_id"`
}

func (e EventAppOpened) EventType() string { return "app_opened" }
func (e EventAppOpened) UserID() string    { return e.ExternalUserID }

// --- Trip Events ---

// EventSearchTrip is sent when a user searches for a trip.
type EventSearchTrip struct {
	ExternalUserID  string  `json:"external_user_id"`
	PickupLocation  string  `json:"pickup_location,omitempty"`
	DropoffLocation string  `json:"dropoff_location,omitempty"`
	PickupLat       float64 `json:"pickup_lat,omitempty"`
	PickupLng       float64 `json:"pickup_lng,omitempty"`
}

func (e EventSearchTrip) EventType() string { return "search_trip" }
func (e EventSearchTrip) UserID() string    { return e.ExternalUserID }

// EventTripBooked is sent when a trip is booked.
type EventTripBooked struct {
	ExternalUserID  string  `json:"external_user_id"`
	ExternalTripID  string  `json:"external_trip_id"`
	PickupLocation  string  `json:"pickup_location,omitempty"`
	DropoffLocation string  `json:"dropoff_location,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
}

func (e EventTripBooked) EventType() string { return "trip_booked" }
func (e EventTripBooked) UserID() string    { return e.ExternalUserID }

// EventTripCompleted is sent when a trip is completed.
type EventTripCompleted struct {
	ExternalUserID  string  `json:"external_user_id"`
	ExternalTripID  string  `json:"external_trip_id"`
	Amount          float64 `json:"amount"`
	PickupLocation  string  `json:"pickup_location,omitempty"`
	DropoffLocation string  `json:"dropoff_location,omitempty"`
}

func (e EventTripCompleted) EventType() string { return "trip_completed" }
func (e EventTripCompleted) UserID() string    { return e.ExternalUserID }

// EventTripCancelled is sent when a trip is cancelled.
type EventTripCancelled struct {
	ExternalUserID string `json:"external_user_id"`
	ExternalTripID string `json:"external_trip_id"`
	CancelReason   string `json:"cancel_reason,omitempty"`
}

func (e EventTripCancelled) EventType() string { return "trip_cancelled" }
func (e EventTripCancelled) UserID() string    { return e.ExternalUserID }

// --- App Behavior Events ---

// EventButtonClicked is sent when a user clicks a button.
type EventButtonClicked struct {
	ExternalUserID string `json:"external_user_id"`
	ButtonID       string `json:"button_id"`
	Screen         string `json:"screen,omitempty"`
}

func (e EventButtonClicked) EventType() string { return "button_clicked" }
func (e EventButtonClicked) UserID() string    { return e.ExternalUserID }

// EventPageViewed is sent when a user views a page/screen.
type EventPageViewed struct {
	ExternalUserID string `json:"external_user_id"`
	PageName       string `json:"page_name"`
	Referrer       string `json:"referrer,omitempty"`
}

func (e EventPageViewed) EventType() string { return "page_viewed" }
func (e EventPageViewed) UserID() string    { return e.ExternalUserID }

// EventEnterLocation is sent when a user enters pickup/dropoff location.
type EventEnterLocation struct {
	ExternalUserID string `json:"external_user_id"`
	Location       string `json:"location"`
	LocationType   string `json:"location_type"` // "pickup" or "dropoff"
}

func (e EventEnterLocation) EventType() string { return "enter_location" }
func (e EventEnterLocation) UserID() string    { return e.ExternalUserID }

// --- Source Constants ---

const (
	SourceOrganic     = "organic"
	SourceFacebookAds = "facebook_ads"
	SourceGoogleAds   = "google_ads"
	SourceTikTokAds   = "tiktok_ads"
	SourceReferral    = "referral"
	SourceB2B         = "b2b"
	SourceEvent       = "event"
)

// --- Device Constants ---

const (
	DeviceIOS     = "ios"
	DeviceAndroid = "android"
	DeviceWeb     = "web"
)
