package order_reject

import (
	"context"

	ucreject "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/reject"
)

type UseCase interface {
	Run(ctx context.Context, in ucreject.In) (*ucreject.Out, error)
}
