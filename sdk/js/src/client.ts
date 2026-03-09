import type { EventType } from './events.js';

export { EventType } from './events.js';

interface EventPayload {
    event_type: string;
    external_user_id: string;
    user_type: string;
    data: Record<string, unknown>;
    timestamp: string;
}

/**
 * CRM SDK Client for tracking events.
 * Reads CRM_API_KEY and CRM_ENDPOINT from environment by default.
 */
export class CRMClient {
    private apiKey: string;
    private endpoint: string;

    constructor(apiKey?: string, endpoint?: string) {
        // Node.js: process.env / Browser: window.__CRM_CONFIG
        const g = globalThis as Record<string, any>;
        const envKey = g.process?.env?.CRM_API_KEY ?? g.__CRM_CONFIG?.apiKey ?? '';
        const envEndpoint = g.process?.env?.CRM_ENDPOINT ?? g.__CRM_CONFIG?.endpoint ?? 'http://localhost:8080';

        this.apiKey = apiKey || envKey;
        this.endpoint = endpoint || envEndpoint;
    }

    /**
     * Track a single event.
     */
    async track(
        eventType: EventType,
        externalUserId: string,
        userType: string,
        data: Record<string, unknown> = {},
    ): Promise<void> {
        const payload: EventPayload = {
            event_type: eventType,
            external_user_id: externalUserId,
            user_type: userType,
            data,
            timestamp: new Date().toISOString(),
        };

        const response = await fetch(`${this.endpoint}/api/v1/ingest/events`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey,
            },
            body: JSON.stringify(payload),
        });

        if (!response.ok) {
            throw new Error(`CRM API returned ${response.status}`);
        }
    }
}
