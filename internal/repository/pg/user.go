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
		TelegramID:  user.TelegramID,
		Username:    user.Username,
		FirstName:   user.FirstName,
		PhoneNumber: user.PhoneNumber,
		Role:        string(domain.RoleBuyer), // Default role for new users
	}

	// GORM Magic: INSERT ... ON CONFLICT (telegram_id) DO UPDATE ...
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "telegram_id"}},
		// Update username and first_name if they changed, but KEEP the existing Role and ID!
		// Also update phone number if provided
		DoUpdates: clause.AssignmentColumns([]string{"username", "first_name", "phone_number", "updated_at"}),
	}).Create(&dbUser).Error

	if err != nil {
		return err
	}

	// Read back the data (to get the UUID and actual Role from the database)
	r.db.WithContext(ctx).Where("telegram_id = ?", user.TelegramID).First(&dbUser)

	// Map back to domain
	user.ID = dbUser.ID
	user.Role = domain.UserRole(dbUser.Role)
	user.PhoneNumber = dbUser.PhoneNumber

	return nil
}

// GetByID fetches a user by UUID.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var dbUser models.User
	if err := r.db.WithContext(ctx).First(&dbUser, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &domain.User{
		ID:          dbUser.ID,
		TelegramID:  dbUser.TelegramID,
		Username:    dbUser.Username,
		FirstName:   dbUser.FirstName,
		PhoneNumber: dbUser.PhoneNumber,
		Role:        domain.UserRole(dbUser.Role),
	}, nil
}

// List retrieves users with filtering and pagination.
func (r *UserRepo) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int64, error) {
	var dbUsers []models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Where("username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR phone_number ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&dbUsers).Error; err != nil {
		return nil, 0, err
	}

	var users []domain.User
	for _, u := range dbUsers {
		users = append(users, domain.User{
			ID:          u.ID,
			TelegramID:  u.TelegramID,
			Username:    u.Username,
			FirstName:   u.FirstName,
			PhoneNumber: u.PhoneNumber,
			Role:        domain.UserRole(u.Role),
		})
	}

	return users, total, nil
}

// UpdateRole changes the role of a user.
func (r *UserRepo) UpdateRole(ctx context.Context, id uuid.UUID, role domain.UserRole) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}

// Delete soft-deletes a user.
func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}
