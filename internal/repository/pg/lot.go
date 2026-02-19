package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
)

type LotRepo struct {
	db *gorm.DB
}

func NewLotRepository(db *gorm.DB) domain.LotRepository {
	return &LotRepo{db: db}
}

// Create inserts a new lot into the PostgreSQL database.
func (r *LotRepo) Create(ctx context.Context, dto *domain.CreateLotDTO) (uuid.UUID, error) {
	// 1. Marshal the strictly typed Params struct into a JSON byte array
	paramsBytes, err := json.Marshal(dto.Params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal lot params: %w", err)
	}

	// 2. Map DTO to GORM Database Model
	dbModel := models.Lot{
		WarehouseID:     dto.WarehouseID,
		Type:            models.LotType(dto.Type),
		Condition:       models.LotCondition(dto.Condition),
		Brand:           dto.Brand,
		Model:           dto.Model,
		Params:          datatypes.JSON(paramsBytes),
		Defects:         dto.Defects,
		Photos:          dto.Photos, // Uses pq.StringArray under the hood
		InitialQuantity: dto.InitialQuantity,
		CurrentQuantity: dto.InitialQuantity, // Initially, current == initial
		PurchasePrice:   dto.PurchasePrice,
		SellPrice:       dto.SellPrice,
		Status:          string(domain.LotStatusActive),
	}

	// 3. Execute the insert query
	if err := r.db.WithContext(ctx).Create(&dbModel).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert lot to db: %w", err)
	}

	return dbModel.ID, nil
}

// ListPublic retrieves a paginated list of lots for Buyers (hides sensitive info and archives).
func (r *LotRepo) ListPublic(ctx context.Context, filter domain.LotFilter) ([]domain.LotPublicResponse, int64, error) {
	var dbModels []models.Lot
	var total int64

	// Buyers only see ACTIVE lots with stock > 0
	query := r.db.WithContext(ctx).Model(&models.Lot{}).
		Where("status = ?", domain.LotStatusActive).
		Where("current_quantity > 0")

	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count public lots: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch public lots: %w", err)
	}

	var responses []domain.LotPublicResponse
	for _, m := range dbModels {
		responses = append(responses, mapToPublicResponse(m))
	}

	return responses, total, nil
}

// ListInternal retrieves a full paginated list of lots for Staff/Admin.
func (r *LotRepo) ListInternal(ctx context.Context, filter domain.LotFilter) ([]domain.LotInternalResponse, int64, error) {
	var dbModels []models.Lot
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Lot{})

	// Staff can filter by any status (including ARCHIVED)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count internal lots: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch internal lots: %w", err)
	}

	var responses []domain.LotInternalResponse
	for _, m := range dbModels {
		internalRes := domain.LotInternalResponse{
			LotPublicResponse: mapToPublicResponse(m),
			WarehouseID:       m.WarehouseID,
			InitialQty:        m.InitialQuantity,
			PurchasePrice:     m.PurchasePrice,
			Status:            m.Status,
		}
		responses = append(responses, internalRes)
	}

	return responses, total, nil
}

// Helper function to apply common search filters
func applyFilters(query *gorm.DB, filter domain.LotFilter) *gorm.DB {
	if filter.Brand != "" {
		query = query.Where("brand ILIKE ?", "%"+filter.Brand+"%")
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	return query
}

// Helper function to map DB model to Public Domain Response
func mapToPublicResponse(m models.Lot) domain.LotPublicResponse {
	var params domain.LotParams
	if len(m.Params) > 0 {
		_ = json.Unmarshal(m.Params, &params)
	}

	return domain.LotPublicResponse{
		ID:              m.ID,
		Type:            string(m.Type),
		Condition:       string(m.Condition),
		Brand:           m.Brand,
		Model:           m.Model,
		Params:          params,
		Defects:         m.Defects,
		Photos:          m.Photos,
		CurrentQuantity: m.CurrentQuantity,
		SellPrice:       m.SellPrice,
	}
}
