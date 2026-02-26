package domain

import (
	"context"

	"github.com/google/uuid"
)

// TransferStatus defines the lifecycle of a warehouse transfer.
type TransferStatus string

const (
	TransferStatusInTransit TransferStatus = "IN_TRANSIT"
	TransferStatusAccepted  TransferStatus = "ACCEPTED"
	TransferStatusCancelled TransferStatus = "CANCELLED"
)

// TransferItemDTO represents a specific lot and quantity to be moved.
type TransferItemDTO struct {
	LotID    uuid.UUID `json:"lot_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"required,gt=0"`
}

// CreateTransferDTO is the payload to initiate a transfer.
type CreateTransferDTO struct {
	FromWarehouseID uuid.UUID         `json:"from_warehouse_id" binding:"required"`
	ToWarehouseID   uuid.UUID         `json:"to_warehouse_id" binding:"required"`
	Items           []TransferItemDTO `json:"items" binding:"required,min=1"`
	Comment         string            `json:"comment"`
}

// TransferRepository handles the complex transactions for moving stock.
type TransferRepository interface {
	CreateTransferTx(ctx context.Context, dto CreateTransferDTO, createdByID uuid.UUID) (uuid.UUID, error)
	AcceptTransferTx(ctx context.Context, transferID uuid.UUID, acceptedByID uuid.UUID) error
}

// TransferService contains the business logic and notifications for transfers.
type TransferService interface {
	CreateTransfer(ctx context.Context, dto CreateTransferDTO, userID uuid.UUID) (uuid.UUID, error)
	AcceptTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
}
