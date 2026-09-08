package cart

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetActiveForUpdate(ctx context.Context, customerID uuid.UUID) (*Cart, error)
	GetActive(ctx context.Context, customerID uuid.UUID) (*Cart, error)
	Create(ctx context.Context, in CreateIn) (*Cart, error)
	GetItems(ctx context.Context, cartID uuid.UUID) ([]Item, error)
	AddItem(ctx context.Context, in AddItemIn) (*Item, error)
	UpdateItem(ctx context.Context, in UpdateItemIn) (*Item, error)
	DeleteItem(ctx context.Context, cartID, itemID uuid.UUID) error
	ClearItems(ctx context.Context, cartID uuid.UUID) error
	Abandon(ctx context.Context, cartID uuid.UUID) error
	MarkConverted(ctx context.Context, cartID, orderID uuid.UUID) error
	CountItems(ctx context.Context, cartID uuid.UUID) (int, error)
	AbandonExpired(ctx context.Context) (int, error)
}

type Cart struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	VenueID    uuid.UUID
	Status     string
	OrderID    *uuid.UUID
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Item struct {
	ID              uuid.UUID
	CartID          uuid.UUID
	MenuItemID      uuid.UUID
	Quantity        int
	SelectedOptions json.RawMessage
	OptionsHash     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateIn struct {
	CustomerID uuid.UUID
	VenueID    uuid.UUID
	ExpiresAt  time.Time
}

type AddItemIn struct {
	CartID          uuid.UUID
	MenuItemID      uuid.UUID
	Quantity        int
	SelectedOptions json.RawMessage
	OptionsHash     string
}

type UpdateItemIn struct {
	CartID   uuid.UUID
	ItemID   uuid.UUID
	Quantity int
}
