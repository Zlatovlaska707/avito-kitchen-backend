package order_ready

import (
	"context"

	uc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/ready"
)

type UseCase interface {
	Run(ctx context.Context, in uc.In) (*uc.Out, error)
}
