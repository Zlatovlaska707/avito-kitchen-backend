package cart_get

import (
	"context"

	cartget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/get"
)

type UseCase interface {
	Run(ctx context.Context, in cartget.In) (*cartget.Out, error)
}
