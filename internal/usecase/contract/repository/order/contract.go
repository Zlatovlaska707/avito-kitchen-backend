package order

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
)

type Repository interface {
	Create(ctx context.Context, in CreateIn) (*Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetByIDForCustomer(ctx context.Context, id, customerID uuid.UUID) (*Order, error)
	GetByIDForVenue(ctx context.Context, id, venueID uuid.UUID) (*Order, error)
	List(ctx context.Context, filter ListFilter) ([]Order, error)
	GetItems(ctx context.Context, orderID uuid.UUID) ([]Item, error)
	GetStatusHistory(ctx context.Context, orderID uuid.UUID) ([]StatusEvent, error)
	ApplyTransition(ctx context.Context, in TransitionIn) (*Order, error)
	FindPlacedBefore(ctx context.Context, before time.Time, limit int) ([]uuid.UUID, error)
}

type Order struct {
	ID               uuid.UUID
	PublicNumber     string
	CustomerID       uuid.UUID
	VenueID          uuid.UUID
	CartID           *uuid.UUID
	Status           orderstatus.Status
	FulfillmentType  orderstatus.FulfillmentType
	DeliveryAddress  json.RawMessage
	ContactPhone     string
	ItemsTotalMinor  int64
	DeliveryFeeMinor int64
	DiscountMinor    int64
	TotalMinor       int64
	Currency         string
	PaymentMethod    string
	PaymentStatus    string
	Comment          *string
	ExternalOrderID  *string
	PrepMinutes      *int
	ETAAt            *time.Time
	PlacedAt         time.Time
	AcceptedAt       *time.Time
	ReadyAt          *time.Time
	DeliveredAt      *time.Time
	ClosedAt         *time.Time
	CancelReason     *string
	CancelledBy      *string
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Item struct {
	ID              uuid.UUID
	OrderID         uuid.UUID
	MenuItemID      *uuid.UUID
	ItemExternalID  string
	NameSnapshot    string
	UnitPriceMinor  int64
	Quantity        int
	OptionsSnapshot json.RawMessage
	LineTotalMinor  int64
	CreatedAt       time.Time
}

type StatusEvent struct {
	FromStatus *orderstatus.Status
	ToStatus   orderstatus.Status
	ActorType  orderstatus.ActorType
	ActorID    *string
	Reason     *string
	CreatedAt  time.Time
}

type CreateIn struct {
	CustomerID       uuid.UUID
	VenueID          uuid.UUID
	CartID           uuid.UUID
	FulfillmentType  orderstatus.FulfillmentType
	DeliveryAddress  json.RawMessage
	ContactPhone     string
	ItemsTotalMinor  int64
	DeliveryFeeMinor int64
	DiscountMinor    int64
	TotalMinor       int64
	PaymentMethod    string
	PaymentStatus    string
	Comment          *string
	PrepMinutes      int
	ETAAt            time.Time
	Items            []CreateItemIn
	OutboxPayload    json.RawMessage
	CreateDelivery   bool
	DeliveryETAMin   int
}

type CreateItemIn struct {
	MenuItemID      *uuid.UUID
	ItemExternalID  string
	NameSnapshot    string
	UnitPriceMinor  int64
	Quantity        int
	OptionsSnapshot json.RawMessage
	LineTotalMinor  int64
}

type TransitionIn struct {
	OrderID         uuid.UUID
	VenueID         *uuid.UUID
	FromStatus      orderstatus.Status
	ToStatus        orderstatus.Status
	ActorType       orderstatus.ActorType
	ActorID         string
	Reason          *string
	CancelledBy     *string
	PrepMinutes     *int
	ETAAt           *time.Time
	FulfillmentType orderstatus.FulfillmentType
	OutboxEventType string
	OutboxPayload   json.RawMessage
}

type ListFilter struct {
	CustomerID *uuid.UUID
	VenueID    *uuid.UUID
	Status     *orderstatus.Status
	Limit      int
	CursorTS   *time.Time
	CursorID   *uuid.UUID
	Offset     int
}
