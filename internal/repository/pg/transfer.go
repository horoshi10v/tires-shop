package pg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
)

type TransferRepo struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) domain.TransferRepository {
	return &TransferRepo{db: db}
}

// CreateTransferTx initiates the transfer and deducts stock from the source warehouse.
func (r *TransferRepo) CreateTransferTx(ctx context.Context, dto domain.CreateTransferDTO, createdByID uuid.UUID) (uuid.UUID, error) {
	var transferID uuid.UUID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var transferItems []models.TransferItem

		for _, item := range dto.Items {
			var lot models.Lot

			// 1. Lock the source lot to prevent concurrent sales/transfers
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lot, "id = ?", item.LotID).Error; err != nil {
				return fmt.Errorf("lot %s not found: %w", item.LotID, err)
			}

			// 2. Validate warehouse ownership and stock
			if lot.WarehouseID != dto.FromWarehouseID {
				return fmt.Errorf("lot %s does not belong to the source warehouse", lot.ID)
			}
			if lot.CurrentQuantity < item.Quantity {
				return fmt.Errorf("not enough stock for lot %s (requested: %d, available: %d)", lot.ID, item.Quantity, lot.CurrentQuantity)
			}

			// 3. Deduct stock and update status if empty
			lot.CurrentQuantity -= item.Quantity
			if lot.CurrentQuantity == 0 {
				lot.Status = string(domain.LotStatusArchived)
			}

			if err := tx.Save(&lot).Error; err != nil {
				return fmt.Errorf("failed to update source lot: %w", err)
			}

			// 4. Prepare transfer item
			transferItems = append(transferItems, models.TransferItem{
				SourceLotID: lot.ID,
				Quantity:    item.Quantity,
			})
		}

		// 5. Create the Transfer document
		transfer := models.Transfer{
			FromWarehouseID: dto.FromWarehouseID,
			ToWarehouseID:   dto.ToWarehouseID,
			Status:          string(domain.TransferStatusInTransit),
			CreatedByID:     createdByID,
			Comment:         dto.Comment,
			Items:           transferItems,
		}

		if err := tx.Create(&transfer).Error; err != nil {
			return fmt.Errorf("failed to create transfer: %w", err)
		}

		transferID = transfer.ID
		return nil
	})

	return transferID, err
}

// AcceptTransferTx completes the transfer by creating new lots at the destination warehouse.
func (r *TransferRepo) AcceptTransferTx(ctx context.Context, transferID uuid.UUID, acceptedByID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var transfer models.Transfer

		// 1. Lock the transfer document
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&transfer, "id = ?", transferID).Error; err != nil {
			return fmt.Errorf("transfer not found: %w", err)
		}

		if transfer.Status != string(domain.TransferStatusInTransit) {
			return fmt.Errorf("transfer is already %s", transfer.Status)
		}

		// 2. Process each item in the transfer
		for i, item := range transfer.Items {
			var sourceLot models.Lot
			if err := tx.First(&sourceLot, "id = ?", item.SourceLotID).Error; err != nil {
				return fmt.Errorf("source lot missing for item %s: %w", item.ID, err)
			}

			// 3. SPLIT LOGIC: Create a new lot at the destination warehouse
			// We copy all metadata (brand, params, price) but set new quantities and warehouse.
			newLot := models.Lot{
				WarehouseID:     transfer.ToWarehouseID,
				Type:            sourceLot.Type,
				Condition:       sourceLot.Condition,
				Brand:           sourceLot.Brand,
				Model:           sourceLot.Model,
				Params:          sourceLot.Params,
				Defects:         sourceLot.Defects,
				Photos:          sourceLot.Photos,
				InitialQuantity: item.Quantity, // The transferred amount becomes the new initial quantity
				CurrentQuantity: item.Quantity,
				PurchasePrice:   sourceLot.PurchasePrice,
				SellPrice:       sourceLot.SellPrice,
				Status:          string(domain.LotStatusActive),
			}

			if err := tx.Create(&newLot).Error; err != nil {
				return fmt.Errorf("failed to create destination lot: %w", err)
			}

			// 4. Link the new lot to the transfer item
			transfer.Items[i].DestinationLotID = &newLot.ID
			if err := tx.Save(&transfer.Items[i]).Error; err != nil {
				return fmt.Errorf("failed to update transfer item: %w", err)
			}
		}

		// 5. Update transfer status
		transfer.Status = string(domain.TransferStatusAccepted)
		transfer.AcceptedByID = &acceptedByID

		if err := tx.Save(&transfer).Error; err != nil {
			return fmt.Errorf("failed to complete transfer: %w", err)
		}

		return nil
	})
}

// CancelTx cancels a pending transfer and returns the stock to the source warehouse.
func (r *TransferRepo) CancelTx(ctx context.Context, transferID uuid.UUID, cancelledByID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var transfer models.Transfer

		// 1. Lock the transfer document
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&transfer, "id = ?", transferID).Error; err != nil {
			return fmt.Errorf("transfer not found: %w", err)
		}

		if transfer.Status != string(domain.TransferStatusInTransit) {
			return fmt.Errorf("cannot cancel transfer with status %s", transfer.Status)
		}

		// 2. Return items to source lot
		for _, item := range transfer.Items {
			var sourceLot models.Lot
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sourceLot, "id = ?", item.SourceLotID).Error; err != nil {
				return fmt.Errorf("source lot missing for item %s: %w", item.ID, err)
			}

			sourceLot.CurrentQuantity += item.Quantity
			// Reactivate lot if it was archived due to 0 stock
			if sourceLot.CurrentQuantity > 0 && sourceLot.Status == string(domain.LotStatusArchived) {
				sourceLot.Status = string(domain.LotStatusActive)
			}

			if err := tx.Save(&sourceLot).Error; err != nil {
				return fmt.Errorf("failed to restock source lot: %w", err)
			}
		}

		// 3. Update transfer status
		transfer.Status = string(domain.TransferStatusCancelled)
		// We could store cancelledByID somewhere if we add a CancelledBy field, but status is enough for now.

		if err := tx.Save(&transfer).Error; err != nil {
			return fmt.Errorf("failed to cancel transfer: %w", err)
		}

		return nil
	})
}

// List retrieves a paginated list of transfers.
func (r *TransferRepo) List(ctx context.Context, filter domain.TransferFilter) ([]domain.TransferResponse, int64, error) {
	var dbTransfers []models.Transfer
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Transfer{}).Preload("Items")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.FromWarehouseID != "" {
		query = query.Where("from_warehouse_id = ?", filter.FromWarehouseID)
	}
	if filter.ToWarehouseID != "" {
		query = query.Where("to_warehouse_id = ?", filter.ToWarehouseID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transfers: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&dbTransfers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch transfers: %w", err)
	}

	var responses []domain.TransferResponse
	for _, t := range dbTransfers {
		var items []domain.TransferItemResponse
		for _, item := range t.Items {
			items = append(items, domain.TransferItemResponse{
				ID:          item.ID,
				SourceLotID: item.SourceLotID,
				Quantity:    item.Quantity,
			})
		}
		responses = append(responses, domain.TransferResponse{
			ID:              t.ID,
			FromWarehouseID: t.FromWarehouseID,
			ToWarehouseID:   t.ToWarehouseID,
			Status:          t.Status,
			CreatedBy:       t.CreatedByID,
			AcceptedBy:      t.AcceptedByID,
			Comment:         t.Comment,
			CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
			Items:           items,
		})
	}

	return responses, total, nil
}

// GetByID retrieves a single transfer by ID.
func (r *TransferRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.TransferResponse, error) {
	var t models.Transfer
	if err := r.db.WithContext(ctx).Preload("Items").First(&t, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("transfer not found: %w", err)
	}

	var items []domain.TransferItemResponse
	for _, item := range t.Items {
		items = append(items, domain.TransferItemResponse{
			ID:          item.ID,
			SourceLotID: item.SourceLotID,
			Quantity:    item.Quantity,
		})
	}

	return &domain.TransferResponse{
		ID:              t.ID,
		FromWarehouseID: t.FromWarehouseID,
		ToWarehouseID:   t.ToWarehouseID,
		Status:          t.Status,
		CreatedBy:       t.CreatedByID,
		AcceptedBy:      t.AcceptedByID,
		Comment:         t.Comment,
		CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
		Items:           items,
	}, nil
}
