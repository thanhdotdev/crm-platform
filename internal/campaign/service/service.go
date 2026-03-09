package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/campaign/domain"
	"github.com/vothanh/crm-platform/internal/campaign/repository"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// CampaignService handles campaign and voucher business logic.
type CampaignService struct {
	campaignRepo repository.CampaignRepository
	voucherRepo  repository.VoucherRepository
	usageRepo    repository.VoucherUsageRepository
}

func NewCampaignService(
	campaignRepo repository.CampaignRepository,
	voucherRepo repository.VoucherRepository,
	usageRepo repository.VoucherUsageRepository,
) *CampaignService {
	return &CampaignService{
		campaignRepo: campaignRepo,
		voucherRepo:  voucherRepo,
		usageRepo:    usageRepo,
	}
}

// CreateCampaign creates a new campaign (per-tenant).
func (s *CampaignService) CreateCampaign(ctx context.Context, c *domain.Campaign) (*domain.Campaign, error) {
	if err := s.campaignRepo.Create(ctx, c); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create campaign", err)
	}
	return c, nil
}

// ListCampaigns returns paginated campaigns for a tenant.
func (s *CampaignService) ListCampaigns(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]domain.Campaign, int64, error) {
	return s.campaignRepo.List(ctx, tenantID, offset, limit)
}

// GetCampaign returns a campaign by ID.
func (s *CampaignService) GetCampaign(ctx context.Context, tenantID, id uuid.UUID) (*domain.Campaign, error) {
	c, err := s.campaignRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch campaign", err)
	}
	if c == nil {
		return nil, apperror.ErrNotFound
	}
	return c, nil
}

// GeneratePolicy50_30_20Voucher creates vouchers based on the 50-30-20 policy.
func (s *CampaignService) GeneratePolicy50_30_20Voucher(ctx context.Context, tenantID, campaignID, customerID uuid.UUID, tripNumber int) (*domain.Voucher, error) {
	discountValue, expiryHours := domain.Policy50_30_20(tripNumber)
	if discountValue == 0 {
		return nil, nil // No voucher for this trip number
	}

	voucher := &domain.Voucher{
		TenantID:      tenantID,
		CampaignID:    campaignID,
		Code:          fmt.Sprintf("TRIP%d-%s", tripNumber, uuid.New().String()[:8]),
		DiscountType:  "fixed",
		DiscountValue: discountValue,
		MaxUsageCount: 1,
		ExpiresAt:     time.Now().Add(time.Duration(expiryHours) * time.Hour),
		IsStackable:   false, // CEO: không cộng dồn
		PeakHourOnly:  true,  // CEO: được dùng vào giờ cao điểm
		TripNumber:    tripNumber,
		IsActive:      true,
	}

	if err := s.voucherRepo.Create(ctx, voucher); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create voucher", err)
	}

	return voucher, nil
}

// RedeemVoucher validates and uses a voucher.
func (s *CampaignService) RedeemVoucher(ctx context.Context, tenantID uuid.UUID, code string, customerID, tripID uuid.UUID) (*domain.Voucher, error) {
	voucher, err := s.voucherRepo.GetByCode(ctx, tenantID, code)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to lookup voucher", err)
	}
	if voucher == nil {
		return nil, apperror.New("VOUCHER_NOT_FOUND", "voucher not found")
	}
	if !voucher.IsActive {
		return nil, apperror.New("VOUCHER_INACTIVE", "voucher is inactive")
	}
	if time.Now().After(voucher.ExpiresAt) {
		return nil, apperror.New("VOUCHER_EXPIRED", "voucher has expired")
	}
	if voucher.UsedCount >= voucher.MaxUsageCount {
		return nil, apperror.New("VOUCHER_USED", "voucher has been fully used")
	}

	// Check if customer already used this voucher
	usageCount, _ := s.usageRepo.CountByCustomerAndVoucher(ctx, customerID, voucher.ID)
	if usageCount > 0 {
		return nil, apperror.New("VOUCHER_ALREADY_USED", "customer already used this voucher")
	}

	// Record usage
	usage := &domain.VoucherUsage{
		TenantID:   tenantID,
		VoucherID:  voucher.ID,
		CustomerID: customerID,
		TripID:     tripID,
		UsedAt:     time.Now(),
	}
	if err := s.usageRepo.Create(ctx, usage); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to record usage", err)
	}

	_ = s.voucherRepo.IncrementUsedCount(ctx, voucher.ID)

	return voucher, nil
}

// ListVouchers returns all vouchers for a campaign.
func (s *CampaignService) ListVouchers(ctx context.Context, tenantID, campaignID uuid.UUID) ([]domain.Voucher, error) {
	return s.voucherRepo.ListByCampaignID(ctx, tenantID, campaignID)
}
