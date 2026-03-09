// Pre-defined event types for type safety
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
    | 'enter_location'
    | (string & {}); // allow custom event types
