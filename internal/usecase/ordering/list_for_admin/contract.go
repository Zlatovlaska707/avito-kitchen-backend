package list_for_admin

import (
	"time"

	"github.com/google/uuid"

	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type In struct {
	VenueID  *uuid.UUID
	Status   *string
	Limit    int
	CursorTS *time.Time
	CursorID *uuid.UUID
	Offset   int
}

type Out struct {
	Orders  []orderrepo.Order
	HasNext bool
	LastTS  *time.Time
	LastID  *uuid.UUID
}
