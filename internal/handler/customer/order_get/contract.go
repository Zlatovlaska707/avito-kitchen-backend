package order_get

import (
	"context"

	orderget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/get_for_customer"
)

type UseCase interface {
	Run(ctx context.Context, in orderget.In) (*orderget.Out, error)
}
