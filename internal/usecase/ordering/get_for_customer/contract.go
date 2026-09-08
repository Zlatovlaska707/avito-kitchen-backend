package get_for_customer

import (
	"github.com/google/uuid"
	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type In struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
}

type Out struct {
	Order    *orderrepo.Order
	Items    []orderrepo.Item
	History  []orderrepo.StatusEvent
	Delivery *deliveryrepo.Delivery
}
