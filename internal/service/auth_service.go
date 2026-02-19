package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/horoshi10v/tires-shop/internal/config"
	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/telegram"
	jwtutil "github.com/horoshi10v/tires-shop/pkg/jwt"
)

type authService struct {
	repo   domain.UserRepository
	cfg    *config.Config
	logger *slog.Logger
}

func NewAuthService(repo domain.UserRepository, cfg *config.Config, logger *slog.Logger) domain.AuthService {
	return &authService{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}
}

// LoginTelegram validates Telegram data, saves the user, and issues a JWT.
func (s *authService) LoginTelegram(ctx context.Context, initData string) (*domain.AuthResponseDTO, error) {
	// 1. Validate cryptographic signature
	tgUser, err := telegram.ValidateInitData(initData, s.cfg.Auth.TelegramBotToken)
	if err != nil {
		s.logger.Warn("invalid telegram login attempt", slog.String("error", err.Error()))
		// For local testing via Postman WITHOUT a real frontend, you can temporarily bypass this
		// by commenting the validation above and hardcoding a tgUser.
		// BUT NEVER DO THIS IN PRODUCTION!
		return nil, fmt.Errorf("authentication failed")
	}

	// 2. Prepare domain user
	domainUser := domain.User{
		TelegramID: tgUser.ID,
		Username:   tgUser.Username,
		FirstName:  tgUser.FirstName,
	}

	// 3. Save or Update user in DB
	if err := s.repo.UpsertUser(ctx, &domainUser); err != nil {
		s.logger.Error("failed to upsert user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("internal server error")
	}

	s.logger.Info("user logged in via telegram", slog.String("username", domainUser.Username), slog.String("role", string(domainUser.Role)))

	// 4. Generate JWT Token
	token, err := jwtutil.GenerateToken(domainUser.ID, string(domainUser.Role), s.cfg.Auth.JWTSecret, s.cfg.Auth.TokenTTL)
	if err != nil {
		s.logger.Error("failed to generate jwt", slog.String("error", err.Error()))
		return nil, fmt.Errorf("internal server error")
	}

	// 5. Return Response
	return &domain.AuthResponseDTO{
		Token: token,
		User:  domainUser,
	}, nil
}
