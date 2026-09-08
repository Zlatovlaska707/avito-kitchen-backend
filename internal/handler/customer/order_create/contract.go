package order_create

import (
	"context"

	ordercreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/create"
)

type UseCase interface {
	Run(ctx context.Context, in ordercreate.In) (*ordercreate.Out, error)
}
