package cancel

import (
	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type In struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	ActorType  orderstatus.ActorType
	ActorID    string
	Reason     *string
}

type Out struct {
	Order *orderrepo.Order
}
