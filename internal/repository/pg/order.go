package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
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
func (r *OrderRepo) CreateOrderTx(ctx context.Context, dto domain.CreateOrderDTO, userID *uuid.UUID) (uuid.UUID, error) {
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
			UserID:        userID, // Link to user if provided
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

// UpdateStatus changes the status of an existing order and records an Audit Log atomically.
func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string, userID uuid.UUID, comment string) error {
	// Start transaction
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order

		// 1. Lock the order row and Preload Items for potential restocking
		if err := tx.Preload("Items").Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, "id = ?", id).Error; err != nil {
			return fmt.Errorf("order not found or locked: %w", err)
		}

		oldStatus := order.Status

		// If status is the same, do nothing
		if oldStatus == newStatus {
			return nil
		}

		// 2. Handle Stock Logic for Cancellations
		if newStatus == "CANCELLED" && oldStatus != "CANCELLED" {
			// Return items to stock
			for _, item := range order.Items {
				var lot models.Lot
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lot, "id = ?", item.LotID).Error; err != nil {
					return fmt.Errorf("lot %s not found during restocking: %w", item.LotID, err)
				}

				lot.CurrentQuantity += item.Quantity
				// Reactivate lot if it was archived due to 0 stock
				if lot.CurrentQuantity > 0 && lot.Status == "ARCHIVED" {
					lot.Status = "ACTIVE"
				}

				if err := tx.Save(&lot).Error; err != nil {
					return fmt.Errorf("failed to restock lot %s: %w", lot.ID, err)
				}
			}
		}

		// Prevent re-opening a cancelled order for safety (simplified logic)
		if oldStatus == "CANCELLED" && newStatus != "CANCELLED" {
			return fmt.Errorf("cannot reopen a cancelled order directly, please create a new order")
		}

		// 3. Update the order
		order.Status = newStatus
		if err := tx.Save(&order).Error; err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		// 4. Prepare old and new values for the Audit Log (as JSON)
		oldVal, _ := json.Marshal(map[string]string{"status": oldStatus})
		newVal, _ := json.Marshal(map[string]string{"status": newStatus})

		// 5. Create the Audit Log record
		auditLog := models.AuditLog{
			Entity:   "ORDER",
			EntityID: order.ID,
			UserID:   userID,
			Action:   "STATUS_CHANGED",
			OldValue: datatypes.JSON(oldVal),
			NewValue: datatypes.JSON(newVal),
			Comment:  comment,
		}

		// 6. Save the Audit Log
		if err := tx.Create(&auditLog).Error; err != nil {
			return fmt.Errorf("failed to write audit log: %w", err)
		}

		return nil // Commit transaction
	})
}

// List retrieves a paginated list of orders with filters.
func (r *OrderRepo) List(ctx context.Context, filter domain.OrderFilter) ([]domain.OrderResponse, int64, error) {
	var dbOrders []models.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Order{}).Preload("Items")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Customer != "" {
		query = query.Where("customer_name ILIKE ? OR customer_phone ILIKE ?", "%"+filter.Customer+"%", "%"+filter.Customer+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&dbOrders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch orders: %w", err)
	}

	var responses []domain.OrderResponse
	for _, order := range dbOrders {
		var items []domain.OrderItemResponse
		for _, item := range order.Items {
			items = append(items, domain.OrderItemResponse{
				LotID:    item.LotID,
				Quantity: item.Quantity,
				Price:    item.PriceAtMoment,
				Total:    item.PriceAtMoment * float64(item.Quantity),
			})
		}

		responses = append(responses, domain.OrderResponse{
			ID:            order.ID,
			CustomerName:  order.CustomerName,
			CustomerPhone: order.CustomerPhone,
			Status:        order.Status,
			TotalAmount:   order.TotalAmount,
			CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
			Items:         items,
		})
	}

	return responses, total, nil
}

// ListByUserID retrieves orders for a specific user.
func (r *OrderRepo) ListByUserID(ctx context.Context, userID uuid.UUID, filter domain.OrderFilter) ([]domain.OrderResponse, int64, error) {
	var dbOrders []models.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Order{}).Where("user_id = ?", userID).Preload("Items")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user orders: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&dbOrders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch user orders: %w", err)
	}

	var responses []domain.OrderResponse
	for _, order := range dbOrders {
		var items []domain.OrderItemResponse
		for _, item := range order.Items {
			items = append(items, domain.OrderItemResponse{
				LotID:    item.LotID,
				Quantity: item.Quantity,
				Price:    item.PriceAtMoment,
				Total:    item.PriceAtMoment * float64(item.Quantity),
			})
		}

		responses = append(responses, domain.OrderResponse{
			ID:            order.ID,
			CustomerName:  order.CustomerName,
			CustomerPhone: order.CustomerPhone,
			Status:        order.Status,
			TotalAmount:   order.TotalAmount,
			CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
			Items:         items,
		})
	}

	return responses, total, nil
}
