package webhook_delete

import (
	"context"

	whdel "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/webhook_delete"
)

type UseCase interface {
	Run(ctx context.Context, in whdel.In) (*whdel.Out, error)
}
