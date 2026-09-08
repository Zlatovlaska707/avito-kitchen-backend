package venue_create

import (
	"context"

	onboard "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/onboard"
)

type UseCase interface {
	Run(ctx context.Context, in onboard.In) (*onboard.Out, error)
}
