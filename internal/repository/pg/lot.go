package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *LotRepo) ListSuggestions(ctx context.Context, filter domain.LotFilter, internal bool, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 8
	}

	var dbModels []models.Lot
	query := r.db.WithContext(ctx).Model(&models.Lot{})
	if !internal {
		query = query.Where("status = ?", domain.LotStatusActive)
	} else if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	query = applyFilters(query, filter)
	query = query.Order("created_at DESC").Limit(24)

	if err := query.Find(&dbModels).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch lot suggestions: %w", err)
	}

	candidateLimit := limit * 4
	if candidateLimit < 24 {
		candidateLimit = 24
	}

	candidates := buildLotSuggestions(dbModels, filter.Search, candidateLimit, nil)
	usageCounts, err := r.loadSuggestionUsage(ctx, candidates, internal)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch suggestion usage: %w", err)
	}

	return buildLotSuggestions(dbModels, filter.Search, limit, usageCounts), nil
}

func (r *LotRepo) TrackSuggestionSelection(ctx context.Context, suggestion string, internal bool) error {
	normalizedSuggestion := normalizeSuggestionKey(suggestion)
	if normalizedSuggestion == "" {
		return nil
	}

	now := time.Now().UTC()
	stat := models.SearchSuggestionStat{
		Scope:                suggestionScope(internal),
		Suggestion:           strings.TrimSpace(suggestion),
		NormalizedSuggestion: normalizedSuggestion,
		UsageCount:           1,
		LastSelectedAt:       now,
	}

	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "scope"}, {Name: "normalized_suggestion"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"suggestion":       stat.Suggestion,
			"usage_count":      gorm.Expr("usage_count + 1"),
			"last_selected_at": now,
			"updated_at":       now,
		}),
	}).Create(&stat).Error; err != nil {
		return fmt.Errorf("failed to track suggestion selection: %w", err)
	}

	return nil
}

func (r *LotRepo) loadSuggestionUsage(ctx context.Context, suggestions []string, internal bool) (map[string]float64, error) {
	normalizedSuggestions := make([]string, 0, len(suggestions))
	seen := make(map[string]struct{}, len(suggestions))
	for _, suggestion := range suggestions {
		key := normalizeSuggestionKey(suggestion)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalizedSuggestions = append(normalizedSuggestions, key)
	}

	if len(normalizedSuggestions) == 0 {
		return map[string]float64{}, nil
	}

	type usageRow struct {
		NormalizedSuggestion string
		UsageCount           int
		LastSelectedAt       time.Time
	}

	var rows []usageRow
	if err := r.db.WithContext(ctx).
		Model(&models.SearchSuggestionStat{}).
		Select("normalized_suggestion, usage_count, last_selected_at").
		Where("scope = ?", suggestionScope(internal)).
		Where("normalized_suggestion IN ?", normalizedSuggestions).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	usageCounts := make(map[string]float64, len(rows))
	now := time.Now().UTC()
	for _, row := range rows {
		usageCounts[row.NormalizedSuggestion] = suggestionUsageBoost(row.UsageCount, row.LastSelectedAt, now)
	}

	return usageCounts, nil
}

func suggestionUsageBoost(usageCount int, lastSelectedAt time.Time, now time.Time) float64 {
	if usageCount <= 0 {
		return 0
	}

	hoursSinceSelection := now.Sub(lastSelectedAt).Hours()
	decayMultiplier := 1.0
	switch {
	case hoursSinceSelection <= 24:
		decayMultiplier = 1.0
	case hoursSinceSelection <= 24*7:
		decayMultiplier = 0.75
	case hoursSinceSelection <= 24*30:
		decayMultiplier = 0.45
	case hoursSinceSelection <= 24*90:
		decayMultiplier = 0.2
	default:
		decayMultiplier = 0.08
	}

	boost := float64(usageCount*25) * decayMultiplier
	if boost > 250 {
		boost = 250
	}

	return boost
}

func suggestionScope(internal bool) string {
	if internal {
		return "STAFF"
	}
	return "PUBLIC"
}

func normalizeSuggestionKey(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func buildLotSuggestions(lots []models.Lot, search string, limit int, usageCounts map[string]float64) []string {
	type rankedSuggestion struct {
		value    string
		score    int
		position int
	}

	parsed := parseStructuredSearch(search)
	queryLower := strings.ToLower(strings.TrimSpace(search))
	freeTextLower := strings.ToLower(strings.TrimSpace(parsed.freeText))
	sizeQuery := ""
	if parsed.width != "" && parsed.profile != "" && parsed.diameter != "" {
		sizeQuery = strings.ToLower(fmt.Sprintf("%s/%s r%s", parsed.width, parsed.profile, parsed.diameter))
	}

	ranked := make([]rankedSuggestion, 0, max(limit, 8))
	seen := make(map[string]struct{}, max(limit, 8))
	position := 0

	appendSuggestion := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		key := normalizeSuggestionKey(value)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}

		score := suggestionScore(value, key, queryLower, freeTextLower, sizeQuery)
		if usageBoost := usageCounts[key]; usageBoost > 0 {
			score += int(usageBoost)
		}
		ranked = append(ranked, rankedSuggestion{value: value, score: score, position: position})
		position++
	}

	for _, lot := range lots {
		var params domain.LotParams
		if len(lot.Params) > 0 {
			_ = json.Unmarshal(lot.Params, &params)
		}
		sizeLabel := buildSuggestionSizeLabel(params, lot.Type)
		yearLabel := ""
		if params.ProductionYear > 0 {
			yearLabel = strconv.Itoa(params.ProductionYear)
		}

		appendSuggestion(sizeLabel)
		appendSuggestion(lot.Brand)
		appendSuggestion(strings.TrimSpace(lot.Brand + " " + lot.Model))
		appendSuggestion(strings.TrimSpace(sizeLabel + " " + lot.Brand))
		appendSuggestion(strings.TrimSpace(sizeLabel + " " + lot.Brand + " " + yearLabel))
		appendSuggestion(yearLabel)

		if lot.Condition == models.ConditionNew {
			appendSuggestion("Новий")
		} else if lot.Condition == models.ConditionUsed {
			appendSuggestion("Вживаний")
		}

		switch params.Season {
		case "SUMMER":
			appendSuggestion("Літо")
		case "WINTER":
			appendSuggestion("Зима")
		case "ALL_SEASON":
			appendSuggestion("Всесезон")
		}

		appendSuggestion(params.CountryOfOrigin)
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].position < ranked[j].position
		}
		return ranked[i].score > ranked[j].score
	})

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	result := make([]string, 0, len(ranked))
	for _, item := range ranked {
		result = append(result, item.value)
	}

	return result
}

func suggestionScore(value string, valueLower string, queryLower string, freeTextLower string, sizeQuery string) int {
	score := 0

	if queryLower == "" {
		switch {
		case strings.Contains(valueLower, "/") && strings.Contains(valueLower, "r"):
			score += 80
		case !strings.Contains(valueLower, " "):
			score += 50
		default:
			score += 20
		}
		return score
	}

	if sizeQuery != "" {
		compactSizeQuery := strings.ReplaceAll(sizeQuery, " ", "")
		compactValue := strings.ReplaceAll(valueLower, " ", "")
		if compactValue == compactSizeQuery {
			score += 200
		} else if strings.HasPrefix(compactValue, compactSizeQuery) {
			score += 140
		} else if strings.Contains(compactValue, compactSizeQuery) {
			score += 100
		}
	}

	if strings.EqualFold(value, queryLower) {
		score += 180
	} else if strings.HasPrefix(valueLower, queryLower) {
		score += 120
	} else if strings.Contains(valueLower, queryLower) {
		score += 70
	}

	if freeTextLower != "" {
		if strings.EqualFold(value, freeTextLower) {
			score += 90
		} else if strings.HasPrefix(valueLower, freeTextLower) {
			score += 60
		} else if strings.Contains(valueLower, freeTextLower) {
			score += 30
		}
	}

	return score
}

func buildSuggestionSizeLabel(params domain.LotParams, lotType models.LotType) string {
	if params.Width > 0 && params.Profile > 0 && params.Diameter > 0 {
		return fmt.Sprintf("%s/%s R%s", formatNumericParam(params.Width), formatNumericParam(params.Profile), formatNumericParam(params.Diameter))
	}
	if lotType == models.LotTypeRim && params.Diameter > 0 {
		return fmt.Sprintf("R%s", formatNumericParam(params.Diameter))
	}
	return ""
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

var tireSizeSearchPattern = regexp.MustCompile(`(?i)(\d{3})\s*/\s*(\d{2,3})\s*r?\s*(\d{2})`)
var tokenPattern = regexp.MustCompile(`[[:alnum:]/.-]+`)

type parsedSearch struct {
	width     string
	profile   string
	diameter  string
	condition string
	season    string
	freeText  string
}

func parseStructuredSearch(value string) parsedSearch {
	result := parsedSearch{}
	remaining := strings.TrimSpace(value)

	if width, profile, diameter, ok := parseSlashTireSizeSearch(remaining); ok {
		result.width = width
		result.profile = profile
		result.diameter = diameter
		remaining = strings.TrimSpace(removeSlashTireSizeToken(remaining))
	}

	tokens := tokenPattern.FindAllString(remaining, -1)
	unusedTokens := make([]string, 0, len(tokens))
	numericTokens := make([]string, 0, 3)

	for _, token := range tokens {
		normalized := strings.ToLower(strings.TrimSpace(token))
		switch {
		case normalized == "new" || strings.HasPrefix(normalized, "нов"):
			result.condition = "NEW"
		case normalized == "used" || strings.HasPrefix(normalized, "вжив"):
			result.condition = "USED"
		case normalized == "summer" || normalized == "літо":
			result.season = "SUMMER"
		case normalized == "winter" || normalized == "зима":
			result.season = "WINTER"
		case normalized == "all-season" || normalized == "allseason" || normalized == "all" || strings.HasPrefix(normalized, "всесез"):
			result.season = "ALL_SEASON"
		case strings.HasPrefix(normalized, "r") && len(normalized) > 1 && isDigits(normalized[1:]):
			if result.diameter == "" {
				result.diameter = normalized[1:]
			} else {
				unusedTokens = append(unusedTokens, token)
			}
		case isDigits(normalized):
			numericTokens = append(numericTokens, normalized)
		default:
			unusedTokens = append(unusedTokens, token)
		}
	}

	if result.width == "" && result.profile == "" && result.diameter == "" && len(numericTokens) >= 3 {
		result.width = numericTokens[0]
		result.profile = numericTokens[1]
		result.diameter = numericTokens[2]
		numericTokens = numericTokens[3:]
	} else if result.width == "" && result.profile == "" && result.diameter != "" && len(numericTokens) >= 2 {
		result.width = numericTokens[0]
		result.profile = numericTokens[1]
		numericTokens = numericTokens[2:]
	}

	unusedTokens = append(unusedTokens, numericTokens...)
	result.freeText = strings.TrimSpace(strings.Join(unusedTokens, " "))

	return result
}

func parseSlashTireSizeSearch(value string) (width string, profile string, diameter string, ok bool) {
	matches := tireSizeSearchPattern.FindStringSubmatch(value)
	if len(matches) != 4 {
		return "", "", "", false
	}

	return matches[1], matches[2], matches[3], true
}

func removeSlashTireSizeToken(value string) string {
	return strings.TrimSpace(tireSizeSearchPattern.ReplaceAllString(value, " "))
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func formatNumericParam(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
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
		parsed := parseStructuredSearch(filter.Search)

		if parsed.width != "" && parsed.profile != "" && parsed.diameter != "" {
			query = query.Where(
				"params->>'width' = ? AND params->>'profile' = ? AND params->>'diameter' = ?",
				parsed.width,
				parsed.profile,
				parsed.diameter,
			)
		}
		if parsed.condition != "" {
			query = query.Where("condition = ?", parsed.condition)
		}
		if parsed.season != "" {
			query = query.Where("params->>'season' = ?", parsed.season)
		}

		if parsed.freeText != "" {
			searchTerm := "%" + parsed.freeText + "%"
			orClauses := []string{
				"brand ILIKE ?",
				"model ILIKE ?",
				"condition::text ILIKE ?",
				"type::text ILIKE ?",
				"params->>'country_of_origin' ILIKE ?",
				"params->>'season' ILIKE ?",
				"concat(coalesce(params->>'width', ''), '/', coalesce(params->>'profile', ''), ' R', coalesce(params->>'diameter', '')) ILIKE ?",
				"concat(coalesce(params->>'width', ''), '/', coalesce(params->>'profile', ''), 'R', coalesce(params->>'diameter', '')) ILIKE ?",
			}
			args := []interface{}{searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm}

			if numberValue, err := strconv.Atoi(parsed.freeText); err == nil {
				orClauses = append(orClauses,
					"params->>'width' = ?",
					"params->>'profile' = ?",
					"params->>'diameter' = ?",
					"params->>'production_year' = ?",
				)
				numberText := strconv.Itoa(numberValue)
				args = append(args, numberText, numberText, numberText, numberText)
			}

			query = query.Where("("+strings.Join(orClauses, " OR ")+")", args...)
		}
	}
	if filter.Width > 0 {
		query = query.Where("params->>'width' = ?", formatNumericParam(filter.Width))
	}
	if filter.Profile > 0 {
		query = query.Where("params->>'profile' = ?", formatNumericParam(filter.Profile))
	}
	if filter.Diameter > 0 {
		query = query.Where("params->>'diameter' = ?", formatNumericParam(filter.Diameter))
	}
	if filter.ProductionYear > 0 {
		query = query.Where("params->>'production_year' = ?", strconv.Itoa(filter.ProductionYear))
	}
	if filter.PCD != "" {
		query = query.Where("params->>'pcd' ILIKE ?", "%"+filter.PCD+"%")
	}
	if filter.DIA > 0 {
		query = query.Where("params->>'dia' = ?", formatNumericParam(filter.DIA))
	}
	if filter.ET != 0 {
		query = query.Where("params->>'et' = ?", formatNumericParam(filter.ET))
	}
	if filter.RimMaterial != "" {
		query = query.Where("params->>'rim_material' = ?", filter.RimMaterial)
	}
	if filter.CountryOfOrigin != "" {
		query = query.Where("params->>'country_of_origin' ILIKE ?", "%"+filter.CountryOfOrigin+"%")
	}
	if filter.Season != "" {
		query = query.Where("params->>'season' = ?", filter.Season)
	}
	if filter.TireTerrain != "" {
		query = query.Where("params->>'tire_terrain' = ?", filter.TireTerrain)
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
		query = query.Where("params->>'ring_inner_diameter' = ?", formatNumericParam(filter.RingInnerDiameter))
	}
	if filter.RingOuterDiameter > 0 {
		query = query.Where("params->>'ring_outer_diameter' = ?", formatNumericParam(filter.RingOuterDiameter))
	}
	if filter.SpacerType != "" {
		query = query.Where("params->>'spacer_type' = ?", filter.SpacerType)
	}
	if filter.SpacerThickness > 0 {
		query = query.Where("params->>'spacer_thickness' = ?", formatNumericParam(filter.SpacerThickness))
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
	if filter.IsCType != nil {
		val := "false"
		if *filter.IsCType {
			val = "true"
		}
		query = query.Where("params->>'is_c_type' = ?", val)
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
