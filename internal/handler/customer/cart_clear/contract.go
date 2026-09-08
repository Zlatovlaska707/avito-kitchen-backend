package cart_clear

import (
	"context"

	cartclear "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/clear"
)

type UseCase interface {
	Run(ctx context.Context, in cartclear.In) (*cartclear.Out, error)
}
