package expire

import (
	"github.com/google/uuid"
)

type In struct {
	OrderID uuid.UUID
}

type Out struct {
	Processed bool
}
