package menu_get

import (
	"context"

	menget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_get"
)

type UseCase interface {
	Run(ctx context.Context, in menget.In) (*menget.Out, error)
}
