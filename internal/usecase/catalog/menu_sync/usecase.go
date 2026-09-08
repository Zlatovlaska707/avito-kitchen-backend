package menu_sync

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type UseCase struct {
	catalog  catalogrepo.Repository
	maxCats  int
	maxItems int
}

func New(catalog catalogrepo.Repository, maxCats, maxItems int) *UseCase {
	return &UseCase{catalog: catalog, maxCats: maxCats, maxItems: maxItems}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	if len(in.Categories) > uc.maxCats {
		return nil, usecase.ErrValidation
	}
	itemCount := 0
	seen := map[string]struct{}{}
	for i := range in.Categories {
		c := &in.Categories[i]
		if _, ok := seen["cat:"+c.ExternalID]; ok {
			return nil, usecase.ErrDuplicateExternalID
		}
		seen["cat:"+c.ExternalID] = struct{}{}
		itemCount += len(c.Items)
		for j := range c.Items {
			if _, ok := seen["item:"+c.Items[j].ExternalID]; ok {
				return nil, usecase.ErrDuplicateExternalID
			}
			seen["item:"+c.Items[j].ExternalID] = struct{}{}
			c.Items[j].CategoryExt = c.ExternalID
		}
	}
	if itemCount > uc.maxItems {
		return nil, usecase.ErrValidation
	}
	ok, err := uc.catalog.TryMenuSyncLock(ctx, in.VenueID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, usecase.ErrMenuSyncInProgress
	}
	result, err := uc.catalog.SyncMenu(ctx, catalogrepo.SyncIn{VenueID: in.VenueID, Categories: in.Categories})
	if err != nil {
		return nil, err
	}
	return &Out{Result: *result}, nil
}
