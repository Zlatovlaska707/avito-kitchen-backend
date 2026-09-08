package reject

import (
	"github.com/google/uuid"

	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type In struct {
	OrderID uuid.UUID
	VenueID uuid.UUID
	Reason  string
}

type Out struct {
	Order *orderrepo.Order
}
