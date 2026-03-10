package repository

import (
	"context"
	"errors"

	"github.com/vothanh/crm-platform/internal/modules/campaign/domain"
	"gorm.io/gorm"
)

type campaignPostgresRepo struct{ db *gorm.DB }

func NewCampaignPostgresRepo(db *gorm.DB) CampaignRepository {
	return &campaignPostgresRepo{db: db}
}

func (r *campaignPostgresRepo) Create(ctx context.Context, c *domain.Campaign) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *campaignPostgresRepo) GetByID(ctx context.Context, tenantID, id uint64) (*domain.Campaign, error) {
	var campaign domain.Campaign
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&campaign).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &campaign, err
}

func (r *campaignPostgresRepo) List(ctx context.Context, tenantID uint64, offset, limit int) ([]domain.Campaign, int64, error) {
	var campaigns []domain.Campaign
	var total int64
	base := r.db.WithContext(ctx).Model(&domain.Campaign{}).Where("tenant_id = ?", tenantID)
	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&campaigns).Error
	return campaigns, total, err
}

func (r *campaignPostgresRepo) Update(ctx context.Context, c *domain.Campaign) error {
	return r.db.WithContext(ctx).Save(c).Error
}

type voucherPostgresRepo struct{ db *gorm.DB }

func NewVoucherPostgresRepo(db *gorm.DB) VoucherRepository {
	return &voucherPostgresRepo{db: db}
}

func (r *voucherPostgresRepo) Create(ctx context.Context, v *domain.Voucher) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *voucherPostgresRepo) GetByCode(ctx context.Context, tenantID uint64, code string) (*domain.Voucher, error) {
	var voucher domain.Voucher
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND code = ?", tenantID, code).First(&voucher).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &voucher, err
}

func (r *voucherPostgresRepo) GetByID(ctx context.Context, tenantID, id uint64) (*domain.Voucher, error) {
	var voucher domain.Voucher
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&voucher).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &voucher, err
}

func (r *voucherPostgresRepo) ListByCampaignID(ctx context.Context, tenantID, campaignID uint64) ([]domain.Voucher, error) {
	var vouchers []domain.Voucher
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND campaign_id = ?", tenantID, campaignID).Find(&vouchers).Error
	return vouchers, err
}

func (r *voucherPostgresRepo) IncrementUsedCount(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&domain.Voucher{}).Where("id = ?", id).
		Update("used_count", gorm.Expr("used_count + 1")).Error
}

type voucherUsagePostgresRepo struct{ db *gorm.DB }

func NewVoucherUsagePostgresRepo(db *gorm.DB) VoucherUsageRepository {
	return &voucherUsagePostgresRepo{db: db}
}

func (r *voucherUsagePostgresRepo) Create(ctx context.Context, vu *domain.VoucherUsage) error {
	return r.db.WithContext(ctx).Create(vu).Error
}

func (r *voucherUsagePostgresRepo) CountByCustomerAndVoucher(ctx context.Context, customerID, voucherID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.VoucherUsage{}).
		Where("customer_id = ? AND voucher_id = ?", customerID, voucherID).Count(&count).Error
	return count, err
}
