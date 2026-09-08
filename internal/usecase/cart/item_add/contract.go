package item_add

import (
	"encoding/json"

	"github.com/google/uuid"
	cartrecalc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/recalc"
)

type In struct {
	CustomerID      uuid.UUID
	VenueID         uuid.UUID
	MenuItemID      uuid.UUID
	Quantity        int
	SelectedOptions json.RawMessage
	OptionIDs       []uuid.UUID
}

type Out struct {
	Result cartrecalc.Result
}
