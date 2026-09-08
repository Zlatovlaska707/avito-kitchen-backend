package create

import (
	"encoding/json"

	"github.com/google/uuid"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type In struct {
	CustomerID         uuid.UUID
	FulfillmentType    string
	DeliveryAddress    json.RawMessage
	ContactPhone       string
	PaymentMethod      string
	Comment            *string
	ExpectedTotalMinor *int64
	IdempotencyKey     string
	RequestHash        string
}

type Out struct {
	Order      *orderrepo.Order
	Replayed   bool
	ReplayBody json.RawMessage
	ReplayCode int
}
