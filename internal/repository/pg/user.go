package pg

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &UserRepo{db: db}
}

// UpsertUser inserts a new user or updates the existing one based on TelegramID.
func (r *UserRepo) UpsertUser(ctx context.Context, user *domain.User) error {
	dbUser := models.User{
		TelegramID: user.TelegramID,
		Username:   user.Username,
		FirstName:  user.FirstName,
		Role:       string(domain.RoleBuyer), // Default role for new users
	}

	// GORM Magic: INSERT ... ON CONFLICT (telegram_id) DO UPDATE ...
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "telegram_id"}},
		// Update username and first_name if they changed, but KEEP the existing Role and ID!
		DoUpdates: clause.AssignmentColumns([]string{"username", "first_name", "updated_at"}),
	}).Create(&dbUser).Error

	if err != nil {
		return err
	}

	// Read back the data (to get the UUID and actual Role from the database)
	r.db.WithContext(ctx).Where("telegram_id = ?", user.TelegramID).First(&dbUser)

	// Map back to domain
	user.ID = dbUser.ID
	user.Role = domain.UserRole(dbUser.Role)

	return nil
}

// GetByID fetches a user by UUID.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var dbUser models.User
	if err := r.db.WithContext(ctx).First(&dbUser, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &domain.User{
		ID:         dbUser.ID,
		TelegramID: dbUser.TelegramID,
		Username:   dbUser.Username,
		FirstName:  dbUser.FirstName,
		Role:       domain.UserRole(dbUser.Role),
	}, nil
}
