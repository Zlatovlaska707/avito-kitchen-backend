package item_update

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	cartrecalc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/recalc"
	nowercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/nower"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type UseCase struct {
	carts   cartrepo.Repository
	catalog catalogrepo.Repository
	venues  venuerepo.Repository
	nower   nowercontract.Nower
	defTZ   *time.Location
}

func New(carts cartrepo.Repository, catalog catalogrepo.Repository, venues venuerepo.Repository, nower nowercontract.Nower, defTZ *time.Location) *UseCase {
	return &UseCase{carts: carts, catalog: catalog, venues: venues, nower: nower, defTZ: defTZ}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	if in.Quantity <= 0 || in.Quantity > 99 {
		return nil, usecase.ErrQuantityLimit
	}
	cart, err := uc.carts.GetActive(ctx, in.CustomerID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrCartNotFound
		}
		return nil, err
	}
	if _, err := uc.carts.UpdateItem(ctx, cartrepo.UpdateItemIn{CartID: cart.ID, ItemID: in.ItemID, Quantity: in.Quantity}); err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrCartNotFound
		}
		return nil, err
	}
	return &Out{Result: uc.recalc(ctx, cart)}, nil
}

func (uc *UseCase) recalc(ctx context.Context, cart *cartrepo.Cart) cartrecalc.Result {
	items, _ := uc.carts.GetItems(ctx, cart.ID)
	ids := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.MenuItemID)
	}
	menuItems, _ := uc.catalog.GetItemsByIDs(ctx, ids)
	venue, _ := uc.venues.GetByID(ctx, cart.VenueID)
	var hours []venuerepo.WorkingHour
	if venue != nil {
		hours, _ = uc.venues.GetWorkingHours(ctx, venue.ID)
	}
	isOpen, reason := shared.VenueIsOpen(venue, hours, uc.nower.Now(), uc.defTZ)
	return cartrecalc.Run(cart, items, menuItems, venue, isOpen, reason)
}
