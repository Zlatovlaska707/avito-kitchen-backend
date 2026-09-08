package accept

import (
	"github.com/google/uuid"

	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type In struct {
	OrderID     uuid.UUID
	VenueID     uuid.UUID
	PrepMinutes *int
}

type Out struct {
	Order *orderrepo.Order
}
