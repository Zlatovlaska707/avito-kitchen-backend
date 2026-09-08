package order_cancel

import (
	"context"

	ordercancel "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/cancel"
)

type UseCase interface {
	Run(ctx context.Context, in ordercancel.In) (*ordercancel.Out, error)
}
