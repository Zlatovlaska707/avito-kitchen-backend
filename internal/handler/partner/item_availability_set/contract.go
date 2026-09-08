package item_availability_set

import (
	"context"

	avail "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/item_availability_set"
)

type UseCase interface {
	Run(ctx context.Context, in avail.In) (*avail.Out, error)
}
