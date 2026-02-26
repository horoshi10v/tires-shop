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
