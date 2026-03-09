package provider

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/notification/domain"
	"gitlab.com/bship1/bship-common-go.git/pkg/zlog"
)

// MockPushProvider is a mock push notification provider for Phase 1.
type MockPushProvider struct{}

func NewMockPushProvider() *MockPushProvider { return &MockPushProvider{} }

func (p *MockPushProvider) Send(ctx context.Context, n *domain.Notification) error {
	zlog.Infof("[MOCK PUSH] To: %s | Title: %s | Content: %s", n.CustomerID, n.Title, n.Content)
	return nil
}

func (p *MockPushProvider) Channel() string { return "push" }

// MockSMSProvider is a mock SMS provider for Phase 1.
type MockSMSProvider struct{}

func NewMockSMSProvider() *MockSMSProvider { return &MockSMSProvider{} }

func (p *MockSMSProvider) Send(ctx context.Context, n *domain.Notification) error {
	zlog.Infof("[MOCK SMS] To: %s | Content: %s", n.CustomerID, n.Content)
	return nil
}

func (p *MockSMSProvider) Channel() string { return "sms" }

// MockEmailProvider is a mock email provider for Phase 1.
type MockEmailProvider struct{}

func NewMockEmailProvider() *MockEmailProvider { return &MockEmailProvider{} }

func (p *MockEmailProvider) Send(ctx context.Context, n *domain.Notification) error {
	zlog.Infof("[MOCK EMAIL] To: %s | Title: %s | Content: %s", n.CustomerID, n.Title, n.Content)
	return nil
}

func (p *MockEmailProvider) Channel() string { return "email" }

// MockZaloProvider is a mock Zalo OA provider for Phase 1.
type MockZaloProvider struct{}

func NewMockZaloProvider() *MockZaloProvider { return &MockZaloProvider{} }

func (p *MockZaloProvider) Send(ctx context.Context, n *domain.Notification) error {
	zlog.Infof("[MOCK ZALO] To: %s | Content: %s", n.CustomerID, n.Content)
	return nil
}

func (p *MockZaloProvider) Channel() string { return "zalo_oa" }
