package clear

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
)

type UseCase struct {
	carts cartrepo.Repository
}

func New(carts cartrepo.Repository) *UseCase {
	return &UseCase{carts: carts}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	cart, err := uc.carts.GetActive(ctx, in.CustomerID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return &Out{}, nil
		}
		return nil, err
	}
	if err := uc.carts.ClearItems(ctx, cart.ID); err != nil {
		return nil, err
	}
	if err := uc.carts.Abandon(ctx, cart.ID); err != nil {
		return nil, err
	}
	return &Out{}, nil
}
