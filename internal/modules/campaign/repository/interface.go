package repository

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/campaign/domain"
)

type CampaignRepository interface {
	Create(ctx context.Context, c *domain.Campaign) error
	GetByID(ctx context.Context, tenantID, id uint64) (*domain.Campaign, error)
	List(ctx context.Context, tenantID uint64, offset, limit int) ([]domain.Campaign, int64, error)
	Update(ctx context.Context, c *domain.Campaign) error
}

type VoucherRepository interface {
	Create(ctx context.Context, v *domain.Voucher) error
	GetByCode(ctx context.Context, tenantID uint64, code string) (*domain.Voucher, error)
	GetByID(ctx context.Context, tenantID, id uint64) (*domain.Voucher, error)
	ListByCampaignID(ctx context.Context, tenantID, campaignID uint64) ([]domain.Voucher, error)
	IncrementUsedCount(ctx context.Context, id uint64) error
}

type VoucherUsageRepository interface {
	Create(ctx context.Context, vu *domain.VoucherUsage) error
	CountByCustomerAndVoucher(ctx context.Context, customerID, voucherID uint64) (int64, error)
}
