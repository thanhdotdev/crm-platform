import type { EventType } from './events';
export { EventType } from './events';
/**
 * CRM SDK Client for tracking events.
 * Reads CRM_API_KEY and CRM_ENDPOINT from environment by default.
 */
export declare class CRMClient {
    private apiKey;
    private endpoint;
    constructor(apiKey?: string, endpoint?: string);
    /**
     * Track a single event.
     */
    track(eventType: EventType, externalUserId: string, userType: string, data?: Record<string, unknown>): Promise<void>;
}
