package menu_sync

import (
	"context"

	menusync "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_sync"
)

type UseCase interface {
	Run(ctx context.Context, in menusync.In) (*menusync.Out, error)
}
