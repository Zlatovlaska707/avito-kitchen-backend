package cart_item_add

import (
	"context"

	cartadd "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_add"
)

type UseCase interface {
	Run(ctx context.Context, in cartadd.In) (*cartadd.Out, error)
}
