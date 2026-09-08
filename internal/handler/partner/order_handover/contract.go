package order_handover

import (
	"context"

	uc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/handover"
)

type UseCase interface {
	Run(ctx context.Context, in uc.In) (*uc.Out, error)
}
