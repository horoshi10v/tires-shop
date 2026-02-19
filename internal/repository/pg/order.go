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

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) domain.OrderRepository {
	return &OrderRepo{db: db}
}

// CreateOrderTx executes order creation and lot deduction atomically.
func (r *OrderRepo) CreateOrderTx(ctx context.Context, dto domain.CreateOrderDTO) (uuid.UUID, error) {
	var orderID uuid.UUID

	// Start a database transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var totalAmount float64
		var orderItems []models.OrderItem

		// 1. Iterate over requested items
		for _, item := range dto.Items {
			var lot models.Lot

			// CRITICAL: SELECT ... FOR UPDATE
			// This locks the row so no other transaction can modify this lot until we are done.
			// This prevents Race Conditions (e.g., selling more tires than available).
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lot, "id = ?", item.LotID).Error; err != nil {
				return fmt.Errorf("lot %s not found: %w", item.LotID, err)
			}

			// 2. Business validation: Check stock
			if lot.CurrentQuantity < item.Quantity {
				return fmt.Errorf("not enough stock for lot %s (requested: %d, available: %d)", lot.ID, item.Quantity, lot.CurrentQuantity)
			}

			// 3. Deduct quantity
			lot.CurrentQuantity -= item.Quantity

			// If lot is empty, we could change its status to 'ARCHIVED' here
			if lot.CurrentQuantity == 0 {
				lot.Status = "ARCHIVED"
			}

			// Save the updated lot
			if err := tx.Save(&lot).Error; err != nil {
				return fmt.Errorf("failed to update lot %s: %w", lot.ID, err)
			}

			// 4. Prepare Order Item
			orderItems = append(orderItems, models.OrderItem{
				LotID:         lot.ID,
				Quantity:      item.Quantity,
				PriceAtMoment: lot.SellPrice,
				CostAtMoment:  lot.PurchasePrice, // Saved securely for P&L reports
			})

			// 5. Accumulate total order amount
			totalAmount += lot.SellPrice * float64(item.Quantity)
		}

		// 6. Create the main Order record
		order := models.Order{
			CustomerName:  dto.CustomerName,
			CustomerPhone: dto.CustomerPhone,
			Status:        "NEW",
			TotalAmount:   totalAmount,
			Items:         orderItems, // GORM will automatically insert these related items
		}

		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		orderID = order.ID
		return nil // Return nil to COMMIT the transaction
	})

	// If err != nil, GORM automatically triggers a ROLLBACK.
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, nil
}