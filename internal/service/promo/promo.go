package promo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"
)

type Repository interface {
	CreatePromo(ctx context.Context, promoCode model.PromoCode) error
	UpdatePromo(ctx context.Context, promoCode model.PromoCode) error
	GetPromo(ctx context.Context, code string) (model.PromoCode, error)
	DeletePromo(ctx context.Context, code string) (model.PromoDeleteResult, error)
	GetTable(ctx context.Context, codes ...string) ([]model.ResponseCode, error)
	GetPromoCode(ctx context.Context, codes ...string) ([]model.StatResponseCode, error)
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
			return fmt.Errorf("failed to create promo: %w", err)
		}
		if err := s.audit.Save(auditLog); err != nil {
			return fmt.Errorf("failed to save audit info: %w", err)
		}
		return nil
	})
}

func (s *Service) UpdatePromoWithAudit(ctx context.Context, modelToRepo model.PromoCode, auditLog audit.Log) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		currentPromo, err := s.repo.GetPromo(ctx, modelToRepo.Code)
		if err != nil {
			return fmt.Errorf("failed to get current promo: %w", err)
		}
		if err := s.repo.UpdatePromo(ctx, modelToRepo); err != nil {
			return fmt.Errorf("failed to update promo: %w", err)
		}
		updatedPromo, err := s.repo.GetPromo(ctx, modelToRepo.Code)
		if err != nil {
			return fmt.Errorf("failed to get updated promo: %w", err)
		}
		auditLog.Changes = promoChanges(currentPromo, updatedPromo)
		if err := s.audit.Save(auditLog); err != nil {
			return fmt.Errorf("failed to save audit info: %w", err)
		}
		return nil
	})
}

func (s *Service) DeletePromoWithAudit(ctx context.Context, code string, auditLog audit.Log) (model.PromoDeleteResult, error) {
	var result model.PromoDeleteResult
	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		deleteResult, err := s.repo.DeletePromo(ctx, code)
		if err != nil {
			return fmt.Errorf("failed to delete promo: %w", err)
		}
		result = deleteResult
		if err := s.audit.Save(auditLog); err != nil {
			return fmt.Errorf("failed to save audit info: %w", err)
		}
		return nil
	})
	return result, err
}

func (s *Service) GetTable(ctx context.Context, codes ...string) ([]model.ResponseCode, error) {
	return s.repo.GetTable(ctx, codes...)
}

func (s *Service) GetStats(ctx context.Context, codes ...string) ([]model.StatResponseCode, error) {
	return s.repo.GetPromoCode(ctx, codes...)
}

func promoChanges(oldPromo, newPromo model.PromoCode) map[string]audit.Change {
	changes := make(map[string]audit.Change)

	if oldPromo.BonusLength != newPromo.BonusLength {
		changes["bonus_length"] = audit.Change{
			Old: strconv.Itoa(oldPromo.BonusLength),
			New: strconv.Itoa(newPromo.BonusLength),
		}
	}
	if !sameDate(oldPromo.Since, newPromo.Since) {
		changes["since"] = audit.Change{
			Old: formatAuditDate(oldPromo.Since),
			New: formatAuditDate(newPromo.Since),
		}
	}
	if !sameDate(oldPromo.Until, newPromo.Until) {
		changes["until"] = audit.Change{
			Old: formatAuditDate(oldPromo.Until),
			New: formatAuditDate(newPromo.Until),
		}
	}
	if oldPromo.Capacity != newPromo.Capacity {
		changes["capacity"] = audit.Change{
			Old: strconv.Itoa(oldPromo.Capacity),
			New: strconv.Itoa(newPromo.Capacity),
		}
	}

	if len(changes) == 0 {
		return nil
	}
	return changes
}

func sameDate(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}

	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func formatAuditDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.DateOnly)
}
