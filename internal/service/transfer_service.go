package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/telegram"
)

type transferService struct {
	repo     domain.TransferRepository
	logger   *slog.Logger
	notifier telegram.Notifier
}

func NewTransferService(repo domain.TransferRepository, logger *slog.Logger, notifier telegram.Notifier) domain.TransferService {
	return &transferService{repo: repo, logger: logger, notifier: notifier}
}

func (s *transferService) CreateTransfer(ctx context.Context, dto domain.CreateTransferDTO, userID uuid.UUID) (uuid.UUID, error) {
	s.logger.Info("initiating warehouse transfer", slog.String("from", dto.FromWarehouseID.String()))

	transferID, err := s.repo.CreateTransferTx(ctx, dto, userID)
	if err != nil {
		s.logger.Error("failed to create transfer", slog.String("error", err.Error()))
		return uuid.Nil, err
	}
	msg := fmt.Sprintf("🚚 Створено переміщення!\nID: %s\nКоментар: %s\nОчікує приймання.", transferID, dto.Comment)
	s.notifier.SendAlert(msg)

	return transferID, nil
}

func (s *transferService) AcceptTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("accepting warehouse transfer", slog.String("transfer_id", transferID.String()))

	if err := s.repo.AcceptTransferTx(ctx, transferID, userID); err != nil {
		s.logger.Error("failed to accept transfer", slog.String("error", err.Error()))
		return err
	}
	msg := fmt.Sprintf("✅ Переміщення ПРИЙНЯТО на склад!\nID: %s\nТовари успішно оприбутковані.", transferID)
	s.notifier.SendAlert(msg)

	return nil
}

func (s *transferService) CancelTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("cancelling warehouse transfer", slog.String("transfer_id", transferID.String()))

	if err := s.repo.CancelTx(ctx, transferID, userID); err != nil {
		s.logger.Error("failed to cancel transfer", slog.String("error", err.Error()))
		return err
	}
	msg := fmt.Sprintf("❌ Переміщення СКАСОВАНО!\nID: %s\nТовари повернуто на склад відправник.", transferID)
	s.notifier.SendAlert(msg)

	return nil
}

func (s *transferService) ListTransfers(ctx context.Context, filter domain.TransferFilter) ([]domain.TransferResponse, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	s.logger.Debug("fetching transfers list", slog.Int("page", filter.Page))
	return s.repo.List(ctx, filter)
}

func (s *transferService) GetTransfer(ctx context.Context, id uuid.UUID) (*domain.TransferResponse, error) {
	s.logger.Debug("fishing transfer details", slog.String("id", id.String()))
	return s.repo.GetByID(ctx, id)
}
