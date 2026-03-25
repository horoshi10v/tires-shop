package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

func (r *LotRepo) Create(ctx context.Context, dto *domain.CreateLotDTO) (uuid.UUID, error) {
	paramsBytes, err := json.Marshal(dto.Params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal lot params: %w", err)
	}

	dbModel := models.Lot{
		WarehouseID:     dto.WarehouseID,
		Type:            models.LotType(dto.Type),
		Condition:       models.LotCondition(dto.Condition),
		Brand:           dto.Brand,
		Model:           dto.Model,
		Params:          datatypes.JSON(paramsBytes),
		Defects:         dto.Defects,
		Photos:          dto.Photos,
		InitialQuantity: dto.InitialQuantity,
		CurrentQuantity: dto.InitialQuantity,
		PurchasePrice:   dto.PurchasePrice,
		SellPrice:       dto.SellPrice,
		Status:          string(domain.LotStatusActive),
	}

	if err := r.db.WithContext(ctx).Create(&dbModel).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert lot to db: %w", err)
	}

	return dbModel.ID, nil
}

func (r *LotRepo) Update(ctx context.Context, id uuid.UUID, dto *domain.UpdateLotDTO) error {
	updates := map[string]interface{}{}

	if dto.WarehouseID != nil {
		updates["warehouse_id"] = *dto.WarehouseID
	}
	if dto.Type != nil {
		updates["type"] = models.LotType(*dto.Type)
	}
	if dto.Condition != nil {
		updates["condition"] = models.LotCondition(*dto.Condition)
	}
	if dto.Brand != nil {
		updates["brand"] = *dto.Brand
	}
	if dto.Model != nil {
		updates["model"] = *dto.Model
	}
	if dto.Params != nil {
		paramsBytes, err := json.Marshal(dto.Params)
		if err != nil {
			return fmt.Errorf("failed to marshal params: %w", err)
		}
		updates["params"] = datatypes.JSON(paramsBytes)
	}
	if dto.Defects != nil {
		updates["defects"] = *dto.Defects
	}
	if dto.Photos != nil {
		updates["photos"] = pq.StringArray(dto.Photos)
	}
	if dto.PurchasePrice != nil {
		updates["purchase_price"] = *dto.PurchasePrice
	}
	if dto.SellPrice != nil {
		updates["sell_price"] = *dto.SellPrice
	}

	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&models.Lot{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update lot: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("lot not found")
	}

	return nil
}

func (r *LotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Lot{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete lot: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("lot not found")
	}
	return nil
}

func (r *LotRepo) ListPublic(ctx context.Context, filter domain.LotFilter) ([]domain.LotPublicResponse, int64, error) {
	var dbModels []models.Lot
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Lot{}).
		Where("status = ?", domain.LotStatusActive)

	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count public lots: %w", err)
	}

	query = applySorting(query, filter)

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch public lots: %w", err)
	}

	responses := make([]domain.LotPublicResponse, 0, len(dbModels))
	for _, m := range dbModels {
		responses = append(responses, mapToPublicResponse(m))
	}

	return responses, total, nil
}

func (r *LotRepo) ListInternal(ctx context.Context, filter domain.LotFilter) ([]domain.LotInternalResponse, int64, error) {
	var dbModels []models.Lot
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Lot{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count internal lots: %w", err)
	}

	query = applySorting(query, filter)

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&dbModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch internal lots: %w", err)
	}

	responses := make([]domain.LotInternalResponse, 0, len(dbModels))
	for _, m := range dbModels {
		responses = append(responses, domain.LotInternalResponse{
			LotPublicResponse: mapToPublicResponse(m),
			WarehouseID:       m.WarehouseID,
			InitialQty:        m.InitialQuantity,
			PurchasePrice:     m.PurchasePrice,
			Status:            m.Status,
		})
	}

	return responses, total, nil
}

func applySorting(query *gorm.DB, filter domain.LotFilter) *gorm.DB {
	sortBy := strings.ToLower(strings.TrimSpace(filter.SortBy))
	sortOrder := strings.ToLower(strings.TrimSpace(filter.SortOrder))
	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	switch sortBy {
	case "price":
		query = query.Order("sell_price " + sortOrder).Order("created_at desc")
	case "stock":
		query = query.
			Order("CASE WHEN current_quantity > 0 THEN 0 ELSE 1 END ASC").
			Order("current_quantity DESC").
			Order("created_at DESC")
	case "popularity":
		query = query.Order("created_at DESC")
	case "created_at", "":
		fallthrough
	default:
		query = query.Order("created_at DESC")
	}

	return query.Order("id DESC")
}

func applyFilters(query *gorm.DB, filter domain.LotFilter) *gorm.DB {
	if filter.Brand != "" {
		query = query.Where("brand ILIKE ?", "%"+filter.Brand+"%")
	}
	if filter.Model != "" {
		query = query.Where("model ILIKE ?", "%"+filter.Model+"%")
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Condition != "" {
		query = query.Where("condition = ?", filter.Condition)
	}
	if filter.CurrentQuantity != nil {
		query = query.Where("current_quantity = ?", *filter.CurrentQuantity)
	}
	if filter.SellPrice != nil {
		query = query.Where("sell_price = ?", *filter.SellPrice)
	}
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Where("brand ILIKE ? OR model ILIKE ?", searchTerm, searchTerm)
	}
	if filter.Width > 0 {
		query = query.Where("params->>'width' = ?", strconv.Itoa(filter.Width))
	}
	if filter.Profile > 0 {
		query = query.Where("params->>'profile' = ?", strconv.Itoa(filter.Profile))
	}
	if filter.Diameter > 0 {
		query = query.Where("params->>'diameter' = ?", strconv.Itoa(filter.Diameter))
	}
	if filter.Season != "" {
		query = query.Where("params->>'season' = ?", filter.Season)
	}
	if filter.AccessoryCategory != "" {
		query = query.Where("params->>'accessory_category' = ?", filter.AccessoryCategory)
	}
	if filter.FastenerType != "" {
		query = query.Where("params->>'fastener_type' = ?", filter.FastenerType)
	}
	if filter.ThreadSize != "" {
		query = query.Where("params->>'thread_size' ILIKE ?", "%"+filter.ThreadSize+"%")
	}
	if filter.SeatType != "" {
		query = query.Where("params->>'seat_type' ILIKE ?", "%"+filter.SeatType+"%")
	}
	if filter.RingInnerDiameter > 0 {
		query = query.Where("params->>'ring_inner_diameter' = ?", strconv.Itoa(filter.RingInnerDiameter))
	}
	if filter.RingOuterDiameter > 0 {
		query = query.Where("params->>'ring_outer_diameter' = ?", strconv.Itoa(filter.RingOuterDiameter))
	}
	if filter.SpacerType != "" {
		query = query.Where("params->>'spacer_type' = ?", filter.SpacerType)
	}
	if filter.SpacerThickness > 0 {
		query = query.Where("params->>'spacer_thickness' = ?", strconv.Itoa(filter.SpacerThickness))
	}
	if filter.PackageQuantity > 0 {
		query = query.Where("params->>'package_quantity' = ?", strconv.Itoa(filter.PackageQuantity))
	}
	if filter.IsRunFlat != nil {
		val := "false"
		if *filter.IsRunFlat {
			val = "true"
		}
		query = query.Where("params->>'is_run_flat' = ?", val)
	}
	if filter.IsSpiked != nil {
		val := "false"
		if *filter.IsSpiked {
			val = "true"
		}
		query = query.Where("params->>'is_spiked' = ?", val)
	}
	if filter.AntiPuncture != nil {
		val := "false"
		if *filter.AntiPuncture {
			val = "true"
		}
		query = query.Where("params->>'anti_puncture' = ?", val)
	}

	return query
}

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
