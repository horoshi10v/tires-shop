package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type userService struct {
	repo   domain.UserRepository
	logger *slog.Logger
}

func NewUserService(repo domain.UserRepository, logger *slog.Logger) domain.UserService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userService) ListUsers(ctx context.Context, filter domain.UserFilter) ([]domain.User, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	s.logger.Debug("fetching users list", slog.Int("page", filter.Page))
	return s.repo.List(ctx, filter)
}

func (s *userService) AddWorker(ctx context.Context, dto domain.CreateWorkerDTO) (uuid.UUID, error) {
	s.logger.Info("attempting to add/promote worker", slog.Any("dto", dto))

	// Case 1: Identify by Telegram ID (Direct Upsert)
	if dto.TelegramID != 0 {
		user := &domain.User{
			TelegramID:  dto.TelegramID,
			Username:    dto.Username,
			FirstName:   dto.FirstName,
			PhoneNumber: dto.PhoneNumber,
			Role:        dto.Role,
		}

		// UpsertUser updates existing or creates new.
		// Note: UpsertUser implementation in Repo currently hardcodes Role=BUYER for new users?
		// We need to verify that logic.
		// Ah, looking at `pg/user.go`:
		// `Role: string(domain.RoleBuyer), // Default role for new users`
		// and `DoUpdates` DOES NOT include `role`.
		// So `UpsertUser` is designed for Login (safe default), not for Admin creation.

		// Strategy: Use UpsertUser to ensure record exists, then explicitly UpdateRole.
		if err := s.repo.UpsertUser(ctx, user); err != nil {
			return uuid.Nil, fmt.Errorf("failed to upsert user for worker creation: %w", err)
		}

		// Now force the role update
		if err := s.repo.UpdateRole(ctx, user.ID, dto.Role); err != nil {
			return uuid.Nil, fmt.Errorf("failed to set role for worker: %w", err)
		}

		return user.ID, nil
	}

	// Case 2: Identify by Username or Phone (Search existing)
	searchTerm := dto.Username
	if searchTerm == "" {
		searchTerm = dto.PhoneNumber
	}

	if searchTerm == "" {
		return uuid.Nil, fmt.Errorf("must provide telegram_id, username, or phone_number")
	}

	// Search for the user
	users, count, err := s.repo.List(ctx, domain.UserFilter{
		Page:     1,
		PageSize: 5,
		Search:   searchTerm,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to search user: %w", err)
	}

	if count == 0 {
		return uuid.Nil, fmt.Errorf("user not found by '%s' (user must interact with bot first if adding by name)", searchTerm)
	}

	// Try to find exact match if multiple results (e.g. search "alex" matches "alexander")
	var targetUser *domain.User
	for _, u := range users {
		if (dto.Username != "" && u.Username == dto.Username) || (dto.PhoneNumber != "" && u.PhoneNumber == dto.PhoneNumber) {
			targetUser = &u
			break
		}
	}

	// If no exact match but only 1 result, assume that's the one
	if targetUser == nil && count == 1 {
		targetUser = &users[0]
	}

	if targetUser == nil {
		return uuid.Nil, fmt.Errorf("ambiguous user match for '%s', please be more specific or use ID", searchTerm)
	}

	// Update role
	if err := s.repo.UpdateRole(ctx, targetUser.ID, dto.Role); err != nil {
		return uuid.Nil, fmt.Errorf("failed to promote user: %w", err)
	}

	s.logger.Info("promoted user to worker", slog.String("username", targetUser.Username), slog.String("role", string(dto.Role)))
	return targetUser.ID, nil
}

func (s *userService) UpdateUserRole(ctx context.Context, id uuid.UUID, role domain.UserRole) error {
	s.logger.Info("updating user role", slog.String("user_id", id.String()), slog.String("role", string(role)))
	return s.repo.UpdateRole(ctx, id, role)
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("deleting user", slog.String("user_id", id.String()))
	return s.repo.Delete(ctx, id)
}
