package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/campaign/domain"
)

type CampaignRepository interface {
	Create(ctx context.Context, c *domain.Campaign) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Campaign, error)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]domain.Campaign, int64, error)
	Update(ctx context.Context, c *domain.Campaign) error
}

type VoucherRepository interface {
	Create(ctx context.Context, v *domain.Voucher) error
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Voucher, error)
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Voucher, error)
	ListByCampaignID(ctx context.Context, tenantID, campaignID uuid.UUID) ([]domain.Voucher, error)
	IncrementUsedCount(ctx context.Context, id uuid.UUID) error
}

type VoucherUsageRepository interface {
	Create(ctx context.Context, vu *domain.VoucherUsage) error
	CountByCustomerAndVoucher(ctx context.Context, customerID, voucherID uuid.UUID) (int64, error)
}
