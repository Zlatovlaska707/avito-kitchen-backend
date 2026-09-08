package advance

import (
	"github.com/google/uuid"
)

type In struct {
	OrderID uuid.UUID
}

type Out struct {
	Completed bool
}
