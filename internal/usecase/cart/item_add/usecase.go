package item_add

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/optionhash"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	cartrecalc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/recalc"
	nowercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/nower"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/errdetails"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type UseCase struct {
	carts    cartrepo.Repository
	catalog  catalogrepo.Repository
	venues   venuerepo.Repository
	nower    nowercontract.Nower
	defTZ    *time.Location
	cartTTL  time.Duration
	maxItems int
}

func New(carts cartrepo.Repository, catalog catalogrepo.Repository, venues venuerepo.Repository, nower nowercontract.Nower, defTZ *time.Location, cartTTL time.Duration, maxItems int) *UseCase {
	return &UseCase{carts: carts, catalog: catalog, venues: venues, nower: nower, defTZ: defTZ, cartTTL: cartTTL, maxItems: maxItems}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	if in.Quantity <= 0 || in.Quantity > 99 {
		return nil, usecase.ErrQuantityLimit
	}
	items, err := uc.catalog.GetItemsByIDs(ctx, []uuid.UUID{in.MenuItemID})
	if err != nil || len(items) == 0 {
		return nil, usecase.ErrMenuItemNotFound
	}
	mi := items[0]
	if !mi.IsActive || !mi.IsAvailable {
		return nil, usecase.ErrMenuItemUnavailable
	}
	if mi.VenueID != in.VenueID {
		return nil, usecase.ErrMenuItemNotFound
	}

	cart, err := uc.carts.GetActive(ctx, in.CustomerID)
	if err != nil {
		if !pgerr.IsNotFound(err) {
			return nil, err
		}
		cart, err = uc.carts.Create(ctx, cartrepo.CreateIn{
			CustomerID: in.CustomerID, VenueID: in.VenueID,
			ExpiresAt: uc.nower.Now().Add(uc.cartTTL),
		})
		if err != nil {
			return nil, err
		}
	} else if cart.VenueID != in.VenueID {
		return nil, fmt.Errorf("%w: %v", usecase.ErrCartVenueMismatch, errdetails.CartVenueMismatchDetail(cart.VenueID))
	}

	count, err := uc.carts.CountItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	if count >= uc.maxItems {
		return nil, usecase.ErrCartItemLimit
	}

	hash := optionhash.Hash(in.OptionIDs)
	selected := in.SelectedOptions
	if len(in.OptionIDs) > 0 {
		_, built, err := shared.BuildOptionsFromIDs(mi, in.OptionIDs)
		if err != nil {
			return nil, err
		}
		selected = built
	} else if len(selected) == 0 {
		selected = json.RawMessage("[]")
	}
	if _, err := uc.carts.AddItem(ctx, cartrepo.AddItemIn{
		CartID: cart.ID, MenuItemID: in.MenuItemID, Quantity: in.Quantity,
		SelectedOptions: selected, OptionsHash: hash,
	}); err != nil {
		return nil, err
	}

	result := uc.recalc(ctx, cart)
	return &Out{Result: result}, nil
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
