package webhook_create

import (
	"context"

	whcreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/webhook_create"
)

type UseCase interface {
	Run(ctx context.Context, in whcreate.In) (*whcreate.Out, error)
}
