package get

import (
	"github.com/google/uuid"
	cartrecalc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/recalc"
)

type In struct {
	CustomerID uuid.UUID
}

type Out struct {
	Result cartrecalc.Result
}
