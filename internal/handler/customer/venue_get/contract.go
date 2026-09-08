package venue_get

import (
	"context"

	venueget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/get"
)

type UseCase interface {
	Run(ctx context.Context, in venueget.In) (*venueget.Out, error)
}
