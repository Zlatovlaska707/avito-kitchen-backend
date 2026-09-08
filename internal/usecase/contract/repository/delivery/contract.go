package delivery

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, in CreateIn) (*Delivery, error)
	UpdateStatus(ctx context.Context, in UpdateStatusIn) (*Delivery, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*Delivery, error)
	FindPendingForAssign(ctx context.Context, before time.Time, limit int) ([]Delivery, error)
	FindAssignedForTransit(ctx context.Context, before time.Time, limit int) ([]Delivery, error)
}

type Delivery struct {
	ID           uuid.UUID
	OrderID      uuid.UUID
	Status       string
	CourierName  *string
	CourierPhone *string
	ETAMinutes   *int
	AssignedAt   *time.Time
	PickedUpAt   *time.Time
	DeliveredAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateIn struct {
	OrderID    uuid.UUID
	ETAMinutes int
}

type UpdateStatusIn struct {
	OrderID      uuid.UUID
	Status       string
	CourierName  *string
	CourierPhone *string
	PickedUpAt   *time.Time
	DeliveredAt  *time.Time
}
