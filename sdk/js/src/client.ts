import type { EventType } from './events';

interface CRMClientConfig {
    apiKey: string;
    endpoint?: string;
}

interface TrackableEvent {
    eventType: EventType;
    externalUserId: string;
    [key: string]: unknown;
}

interface EventPayload {
    event_type: string;
    external_user_id: string;
    data: Record<string, unknown>;
    timestamp: string;
}

/**
 * CRM SDK Client for tracking events from client apps (React Native / Web).
 * Uses a **client key** (`ck_live_*`) which is safe to embed in client code.
 */
export class CRMClient {
    private apiKey: string;
    private endpoint: string;
    private queue: EventPayload[] = [];
    private flushTimer: ReturnType<typeof setTimeout> | null = null;
    private batchSize: number;
    private flushInterval: number;

    constructor(config: CRMClientConfig) {
        this.apiKey = config.apiKey;
        this.endpoint = config.endpoint || 'http://localhost:8080';
        this.batchSize = 10;
        this.flushInterval = 5000; // 5 seconds
    }

    /**
     * Track a single event. Events are queued and flushed in batches.
     */
    track(event: TrackableEvent): void {
        const { eventType, externalUserId, ...data } = event;

        // Convert camelCase to snake_case for API
        const snakeData: Record<string, unknown> = {};
        for (const [key, value] of Object.entries(data)) {
            snakeData[camelToSnake(key)] = value;
        }

        this.queue.push({
            event_type: eventType,
            external_user_id: externalUserId,
            data: snakeData,
            timestamp: new Date().toISOString(),
        });

        if (this.queue.length >= this.batchSize) {
            this.flush();
        } else if (!this.flushTimer) {
            this.flushTimer = setTimeout(() => this.flush(), this.flushInterval);
        }
    }

    /**
     * Immediately send all queued events.
     */
    async flush(): Promise<void> {
        if (this.flushTimer) {
            clearTimeout(this.flushTimer);
            this.flushTimer = null;
        }

        if (this.queue.length === 0) return;

        const events = [...this.queue];
        this.queue = [];

        try {
            if (events.length === 1) {
                await this.send('/api/v1/ingest/events', events[0]);
            } else {
                await this.send('/api/v1/ingest/events/batch', { events });
            }
        } catch (err) {
            // Re-queue on failure (with limit to prevent memory leak)
            if (this.queue.length < 100) {
                this.queue.push(...events);
            }
            console.error('[CRM SDK] Failed to send events:', err);
        }
    }

    private async send(path: string, body: unknown): Promise<void> {
        const response = await fetch(`${this.endpoint}${path}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey,
            },
            body: JSON.stringify(body),
        });

        if (!response.ok) {
            throw new Error(`CRM API returned ${response.status}`);
        }
    }
}

function camelToSnake(str: string): string {
    return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);
}
