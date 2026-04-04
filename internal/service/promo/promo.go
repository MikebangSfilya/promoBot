package promo

import (
	"context"
	"fmt"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"
)

type Repository interface {
	CreatePromo(ctx context.Context, promoCode model.PromoCode) error
	GetTable(ctx context.Context) ([]model.ResponseCode, error)
	GetPromoCode(ctx context.Context, codes []string) ([]model.StatResponseCode, error)
}

type AuditSaver interface {
	Save(s audit.Log) error
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	repo      Repository
	audit     AuditSaver
	txManager TxManager
}

func NewSaveService(repo Repository, audit AuditSaver, tx TxManager) *Service {
	return &Service{
		repo:      repo,
		audit:     audit,
		txManager: tx,
	}
}

func (s *Service) CreatePromoWithAudit(ctx context.Context, modelToRepo model.PromoCode, auditLog audit.Log) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.CreatePromo(ctx, modelToRepo); err != nil {
			return fmt.Errorf("failed to create modelToRepo: %w", err)
		}
		if err := s.audit.Save(auditLog); err != nil {
			return fmt.Errorf("failed to save audit info: %w", err)
		}
		return nil
	})
}

func (s *Service) GetTable(ctx context.Context) ([]model.ResponseCode, error) {
	return s.repo.GetTable(ctx)
}

func (s *Service) GetStats(ctx context.Context, codes []string) ([]model.StatResponseCode, error) {
	return s.repo.GetPromoCode(ctx, codes)
}
