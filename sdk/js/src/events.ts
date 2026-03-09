// Event types — pre-defined for type safety
export interface BaseEvent {
    externalUserId: string;
}

// --- User Lifecycle ---

export interface UserRegisteredEvent extends BaseEvent {
    phone?: string;
    fullName?: string;
    email?: string;
    source?: string;
    campaignId?: string;
    deviceType?: string;
}

export interface AppInstalledEvent extends BaseEvent {
    deviceType?: string;
}

export interface AppOpenedEvent extends BaseEvent { }

// --- Trip ---

export interface SearchTripEvent extends BaseEvent {
    pickupLocation?: string;
    dropoffLocation?: string;
    pickupLat?: number;
    pickupLng?: number;
}

export interface TripBookedEvent extends BaseEvent {
    externalTripId: string;
    pickupLocation?: string;
    dropoffLocation?: string;
    amount?: number;
}

export interface TripCompletedEvent extends BaseEvent {
    externalTripId: string;
    amount: number;
    pickupLocation?: string;
    dropoffLocation?: string;
}

export interface TripCancelledEvent extends BaseEvent {
    externalTripId: string;
    cancelReason?: string;
}

// --- App Behavior ---

export interface ButtonClickedEvent extends BaseEvent {
    buttonId: string;
    screen?: string;
}

export interface PageViewedEvent extends BaseEvent {
    pageName: string;
    referrer?: string;
}

export interface EnterLocationEvent extends BaseEvent {
    location: string;
    locationType: 'pickup' | 'dropoff';
}

// Event type enum
export type EventType =
    | 'user_registered'
    | 'app_installed'
    | 'app_opened'
    | 'search_trip'
    | 'trip_booked'
    | 'trip_completed'
    | 'trip_cancelled'
    | 'button_clicked'
    | 'page_viewed'
    | 'enter_location';

// Factory functions
export const Events = {
    UserRegistered: (data: UserRegisteredEvent) =>
        ({ eventType: 'user_registered' as const, ...data }),
    AppInstalled: (data: AppInstalledEvent) =>
        ({ eventType: 'app_installed' as const, ...data }),
    AppOpened: (data: AppOpenedEvent) =>
        ({ eventType: 'app_opened' as const, ...data }),
    SearchTrip: (data: SearchTripEvent) =>
        ({ eventType: 'search_trip' as const, ...data }),
    TripBooked: (data: TripBookedEvent) =>
        ({ eventType: 'trip_booked' as const, ...data }),
    TripCompleted: (data: TripCompletedEvent) =>
        ({ eventType: 'trip_completed' as const, ...data }),
    TripCancelled: (data: TripCancelledEvent) =>
        ({ eventType: 'trip_cancelled' as const, ...data }),
    ButtonClicked: (data: ButtonClickedEvent) =>
        ({ eventType: 'button_clicked' as const, ...data }),
    PageViewed: (data: PageViewedEvent) =>
        ({ eventType: 'page_viewed' as const, ...data }),
    EnterLocation: (data: EnterLocationEvent) =>
        ({ eventType: 'enter_location' as const, ...data }),
};

// Source constants
export const Source = {
    Organic: 'organic',
    FacebookAds: 'facebook_ads',
    GoogleAds: 'google_ads',
    TikTokAds: 'tiktok_ads',
    Referral: 'referral',
    B2B: 'b2b',
    Event: 'event',
} as const;

// Device constants
export const Device = {
    iOS: 'ios',
    Android: 'android',
    Web: 'web',
} as const;
