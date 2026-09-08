package order_accept

import (
	"context"

	uc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/accept"
)

type UseCase interface {
	Run(ctx context.Context, in uc.In) (*uc.Out, error)
}
