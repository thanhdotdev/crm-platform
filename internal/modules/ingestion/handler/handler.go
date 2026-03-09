package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	customerDomain "github.com/vothanh/crm-platform/internal/modules/customer/domain"
	customerService "github.com/vothanh/crm-platform/internal/modules/customer/service"
	tripDomain "github.com/vothanh/crm-platform/internal/modules/trip/domain"
	tripService "github.com/vothanh/crm-platform/internal/modules/trip/service"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/response"
	"gorm.io/datatypes"
)

// IngestionHandler receives events from the SDK and routes them to appropriate modules.
type IngestionHandler struct {
	customerSvc *customerService.CustomerService
	tripSvc     *tripService.TripService
}

// NewIngestionHandler creates a new IngestionHandler.
func NewIngestionHandler(
	customerSvc *customerService.CustomerService,
	tripSvc *tripService.TripService,
) *IngestionHandler {
	return &IngestionHandler{
		customerSvc: customerSvc,
		tripSvc:     tripSvc,
	}
}

// RegisterRoutes registers ingestion routes (protected by API key auth).
func (h *IngestionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/ingest/events", h.IngestEvent)
}

type ingestRequest struct {
	EventType      string          `json:"event_type" binding:"required"`
	ExternalUserID string          `json:"external_user_id" binding:"required"`
	UserType       string          `json:"user_type" binding:"required"`
	Data           json.RawMessage `json:"data"`
	Timestamp      *time.Time      `json:"timestamp"`
}

func (h *IngestionHandler) IngestEvent(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	var req ingestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.processEvent(c, tenantID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, result)
}

// processEvent routes the event to the appropriate module.
func (h *IngestionHandler) processEvent(c *gin.Context, tenantID uuid.UUID, req *ingestRequest) (interface{}, error) {
	ctx := c.Request.Context()
	keyType := middleware.GetKeyType(c)

	// First, ensure customer exists (upsert)
	customer, err := h.ensureCustomer(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	log.Printf("Customer found: %v", customer)

	// Record the event in timeline
	eventData := datatypes.JSON(req.Data)
	event := &customerDomain.EventLog{
		TenantID:  tenantID,
		UserID:    customer.ID,
		UserType:  req.UserType,
		EventType: req.EventType,
		EventData: eventData,
		Source:    "sdk_" + keyType,
		EventTime: timeOrNow(req.Timestamp),
	}
	_ = h.customerSvc.RecordEvent(ctx, event)

	// Route to specific module based on event type
	// switch req.EventType {
	// case "user_registered", "app_installed":
	// 	return h.handleUserEvent(ctx, tenantID, customer, req)
	// case "trip_booked":
	// 	return h.handleTripBooked(ctx, tenantID, customer, req)
	// case "trip_completed":
	// 	return h.handleTripCompleted(ctx, tenantID, customer, req)
	// case "trip_cancelled":
	// 	return h.handleTripCancelled(ctx, tenantID, customer, req)
	// default:
	// 	// Behavioral events (app_opened, button_clicked, etc.) — just recorded above
	// 	return map[string]interface{}{
	// 		"customer_id": customer.ID,
	// 		"event_type":  req.EventType,
	// 		"recorded":    true,
	// 	}, nil
	// }

	return map[string]interface{}{
		"customer_id": nil,
		"event_type":  req.EventType,
		"recorded":    true,
	}, nil
}

// ensureCustomer finds or creates a customer from the event.
func (h *IngestionHandler) ensureCustomer(ctx context.Context, tenantID uuid.UUID, req *ingestRequest) (*customerDomain.Customer, error) {
	customer := &customerDomain.Customer{
		TenantID:   tenantID,
		ExternalID: req.ExternalUserID,
	}

	// Parse additional fields from data
	var data map[string]interface{}
	if req.Data != nil {
		_ = json.Unmarshal(req.Data, &data)
	}

	if v, ok := data["phone"].(string); ok {
		customer.Phone = v
	}
	if v, ok := data["full_name"].(string); ok {
		customer.FullName = v
	}
	if v, ok := data["email"].(string); ok {
		customer.Email = v
	}
	if v, ok := data["source"].(string); ok {
		customer.Source = v
	}
	if v, ok := data["campaign_id"].(string); ok {
		customer.CampaignID = v
	}
	if v, ok := data["device_type"].(string); ok {
		customer.DeviceType = v
	}

	if req.EventType == "app_installed" {
		now := time.Now()
		customer.AppInstalledAt = &now
	}

	return h.customerSvc.UpsertCustomer(ctx, customer)
}

func (h *IngestionHandler) handleUserEvent(ctx context.Context, tenantID uuid.UUID, customer *customerDomain.Customer, req *ingestRequest) (interface{}, error) {
	return map[string]interface{}{
		"customer_id":     customer.ID,
		"lifecycle_stage": customer.LifecycleStage,
		"lead_score":      customer.LeadScore,
	}, nil
}

func (h *IngestionHandler) handleTripBooked(ctx context.Context, tenantID uuid.UUID, customer *customerDomain.Customer, req *ingestRequest) (interface{}, error) {
	var data map[string]interface{}
	if req.Data != nil {
		_ = json.Unmarshal(req.Data, &data)
	}

	trip := &tripDomain.Trip{
		TenantID:   tenantID,
		CustomerID: customer.ID,
		Status:     tripDomain.TripStatusBooked,
		BookedAt:   timeOrNow(req.Timestamp),
	}

	if v, ok := data["external_trip_id"].(string); ok {
		trip.ExternalTripID = v
	}
	if v, ok := data["pickup_location"].(string); ok {
		trip.PickupLocation = v
	}
	if v, ok := data["dropoff_location"].(string); ok {
		trip.DropoffLocation = v
	}
	if v, ok := data["amount"].(float64); ok {
		trip.Amount = v
	}

	// trip_number is NOT assigned at booking — only on completion
	created, err := h.tripSvc.CreateTrip(ctx, trip)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"trip_id": created.ID,
		"status":  created.Status,
	}, nil
}

func (h *IngestionHandler) handleTripCompleted(ctx context.Context, tenantID uuid.UUID, customer *customerDomain.Customer, req *ingestRequest) (interface{}, error) {
	var data map[string]interface{}
	if req.Data != nil {
		_ = json.Unmarshal(req.Data, &data)
	}

	var amount float64
	if v, ok := data["amount"].(float64); ok {
		amount = v
	}

	// Update customer stats
	luxuryUpgraded, _ := h.customerSvc.IncrementTrip(ctx, tenantID, customer.ID, amount)

	// If external trip ID provided, create/update the trip record
	if extID, ok := data["external_trip_id"].(string); ok && extID != "" {
		trip := &tripDomain.Trip{
			TenantID:       tenantID,
			CustomerID:     customer.ID,
			ExternalTripID: extID,
			Status:         tripDomain.TripStatusCompleted,
			Amount:         amount,
			BookedAt:       timeOrNow(req.Timestamp),
		}
		now := time.Now()
		trip.CompletedAt = &now
		created, err := h.tripSvc.CreateTrip(ctx, trip)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"trip_id":         created.ID,
			"luxury_upgraded": luxuryUpgraded,
		}, nil
	}

	// No external trip ID — stats already updated
	return map[string]interface{}{
		"customer_id":     customer.ID,
		"luxury_upgraded": luxuryUpgraded,
	}, nil
}

func (h *IngestionHandler) handleTripCancelled(ctx context.Context, tenantID uuid.UUID, customer *customerDomain.Customer, req *ingestRequest) (interface{}, error) {
	var data map[string]interface{}
	if req.Data != nil {
		_ = json.Unmarshal(req.Data, &data)
	}

	if extID, ok := data["external_trip_id"].(string); ok && extID != "" {
		trip := &tripDomain.Trip{
			TenantID:       tenantID,
			CustomerID:     customer.ID,
			ExternalTripID: extID,
			Status:         tripDomain.TripStatusCancelled,
			BookedAt:       timeOrNow(req.Timestamp),
		}
		if v, ok := data["cancel_reason"].(string); ok {
			trip.CancelReason = v
		}
		created, err := h.tripSvc.CreateTrip(ctx, trip)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"trip_id":       created.ID,
			"status":        "cancelled",
			"cancel_reason": trip.CancelReason,
		}, nil
	}

	return map[string]interface{}{
		"customer_id": customer.ID,
		"event_type":  "trip_cancelled",
	}, nil
}

func timeOrNow(t *time.Time) time.Time {
	if t != nil {
		return *t
	}
	return time.Now()
}
