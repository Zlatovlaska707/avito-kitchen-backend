package assign

import (
	"github.com/google/uuid"
)

type In struct {
	OrderID      uuid.UUID
	CourierName  string
	CourierPhone string
}

type Out struct {
	Assigned bool
}
