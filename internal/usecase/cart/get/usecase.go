package get

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
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
	cart, err := uc.carts.GetActive(ctx, in.CustomerID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return &Out{Result: cartrecalc.Result{}}, nil
		}
		return nil, err
	}
	return &Out{Result: uc.recalc(ctx, cart)}, nil
}

func (uc *UseCase) recalc(ctx context.Context, cart *cartrepo.Cart) cartrecalc.Result {
	items, err := uc.carts.GetItems(ctx, cart.ID)
	if err != nil {
		return cartrecalc.Result{Cart: cart}
	}
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
