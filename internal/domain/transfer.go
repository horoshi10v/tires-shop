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

// TransferFilter defines criteria for searching transfers.
type TransferFilter struct {
	Page            int
	PageSize        int
	Status          string
	FromWarehouseID string
	ToWarehouseID   string
}

// TransferResponse represents the transfer data returned to the client.
type TransferResponse struct {
	ID              uuid.UUID              `json:"id"`
	FromWarehouseID uuid.UUID              `json:"from_warehouse_id"`
	ToWarehouseID   uuid.UUID              `json:"to_warehouse_id"`
	Status          string                 `json:"status"`
	CreatedBy       uuid.UUID              `json:"created_by"`
	AcceptedBy      *uuid.UUID             `json:"accepted_by,omitempty"`
	Comment         string                 `json:"comment"`
	CreatedAt       string                 `json:"created_at"`
	Items           []TransferItemResponse `json:"items,omitempty"`
}

// TransferItemResponse represents a single item in the transfer response.
type TransferItemResponse struct {
	ID          uuid.UUID `json:"id"`
	SourceLotID uuid.UUID `json:"source_lot_id"`
	Quantity    int       `json:"quantity"`
}

// TransferRepository handles the complex transactions for moving stock.
type TransferRepository interface {
	CreateTransferTx(ctx context.Context, dto CreateTransferDTO, createdByID uuid.UUID) (uuid.UUID, error)
	AcceptTransferTx(ctx context.Context, transferID uuid.UUID, acceptedByID uuid.UUID) error
	CancelTx(ctx context.Context, transferID uuid.UUID, cancelledByID uuid.UUID) error
	List(ctx context.Context, filter TransferFilter) ([]TransferResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*TransferResponse, error)
}

// TransferService contains the business logic and notifications for transfers.
type TransferService interface {
	CreateTransfer(ctx context.Context, dto CreateTransferDTO, userID uuid.UUID) (uuid.UUID, error)
	AcceptTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
	CancelTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
	ListTransfers(ctx context.Context, filter TransferFilter) ([]TransferResponse, int64, error)
	GetTransfer(ctx context.Context, id uuid.UUID) (*TransferResponse, error)
}
