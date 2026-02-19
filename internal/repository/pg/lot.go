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
	// 1. Marshal the flexible params map into a JSON byte array
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
		InitialQuantity: dto.InitialQuantity,
		CurrentQuantity: dto.InitialQuantity, // Initially, current == initial
		PurchasePrice:   dto.PurchasePrice,
		SellPrice:       dto.SellPrice,
		Status:          "ACTIVE", // Default status for a new lot
	}

	// 3. Execute the insert query
	if err := r.db.WithContext(ctx).Create(&dbModel).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert lot to db: %w", err)
	}

	return dbModel.ID, nil
}

// List retrieves a paginated and filtered list of lots.
func (r *LotRepo) List(ctx context.Context, filter domain.LotFilter) ([]domain.LotResponse, int64, error) {
	var dbModels []models.Lot
	var total int64

	// 1. Start building the query
	query := r.db.WithContext(ctx).Model(&models.Lot{})

	// 2. Apply filters dynamically
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Brand != "" {
		// ILIKE makes the search case-insensitive in PostgreSQL
		query = query.Where("brand ILIKE ?", "%"+filter.Brand+"%")
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	// 3. Count total records BEFORE applying limit/offset (needed for frontend pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count lots: %w", err)
	}

	// 4. Apply Pagination (Offset and Limit)
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch lots: %w", err)
	}

	// 5. Map DB Models to Domain Responses
	var responses []domain.LotResponse
	for _, m := range dbModels {
		var params map[string]interface{}
		// Unmarshal the JSONB params back to a Go map
		if len(m.Params) > 0 {
			_ = json.Unmarshal(m.Params, &params)
		}

		responses = append(responses, domain.LotResponse{
			ID:              m.ID,
			WarehouseID:     m.WarehouseID,
			Type:            string(m.Type),
			Condition:       string(m.Condition),
			Brand:           m.Brand,
			Model:           m.Model,
			Params:          params,
			CurrentQuantity: m.CurrentQuantity,
			SellPrice:       m.SellPrice,
			Status:          m.Status,
		})
	}

	return responses, total, nil
}