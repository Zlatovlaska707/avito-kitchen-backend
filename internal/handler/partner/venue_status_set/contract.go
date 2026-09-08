package venue_status_set

import (
	"context"

	statusset "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/status_set"
)

type UseCase interface {
	Run(ctx context.Context, in statusset.In) (*statusset.Out, error)
}
