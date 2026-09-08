package cart_item_update

import (
	"context"

	cartupdate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_update"
)

type UseCase interface {
	Run(ctx context.Context, in cartupdate.In) (*cartupdate.Out, error)
}
