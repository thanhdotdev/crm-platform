import { CRMClient } from '../src/client.js';

async function main() {
    console.log("Initializing CRM JS SDK Client...");

    // Ensure CRM_API_KEY is available in process.env (passed from test_sdk.sh)
    if (!process.env.CRM_API_KEY) {
        console.error("❌ CRM_API_KEY is not set. Please run via scripts/test_sdk.sh");
        process.exit(1);
    }

    const crm = new CRMClient();

    console.log("Sending test event: 'trip_completed'");
    try {
        await crm.track(
            "trip_completed",
            "test-user-002",
            "driver",
            {
                amount: 500000,
                external_trip_id: "trip-js-001",
                rating: 5
            }
        );
        console.log("✅ Successfully sent event via JS SDK!");
    } catch (err) {
        console.error("❌ Failed to send event via JS SDK:", err);
        process.exit(1);
    }
}

main();
