package item_update

import (
	"github.com/google/uuid"
	cartrecalc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/recalc"
)

type In struct {
	CustomerID uuid.UUID
	ItemID     uuid.UUID
	Quantity   int
}

type Out struct {
	Result cartrecalc.Result
}
