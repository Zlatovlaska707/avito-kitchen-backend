package create

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	trm "github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/geo"
	"github.com/zlatovlaska707/avito-kitchen/internal/domain/money"
	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	nowercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/nower"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	idempotencyrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/idempotency"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/errdetails"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type UseCase struct {
	trm           trm.Manager
	carts         cartrepo.Repository
	catalog       catalogrepo.Repository
	venues        venuerepo.Repository
	orders        orderrepo.Repository
	idempotency   idempotencyrepo.Repository
	nower         nowercontract.Nower
	defTZ         *time.Location
	idemTTL       time.Duration
	deliveryBase  int
	deliveryPerKm float64
}

func New(trmMgr trm.Manager, carts cartrepo.Repository, catalog catalogrepo.Repository, venues venuerepo.Repository,
	orders orderrepo.Repository, idempotency idempotencyrepo.Repository, nower nowercontract.Nower, defTZ *time.Location,
	idemTTL time.Duration, deliveryBase int, deliveryPerKm float64,
) *UseCase {
	return &UseCase{
		trm: trmMgr, carts: carts, catalog: catalog, venues: venues, orders: orders,
		idempotency: idempotency, nower: nower, defTZ: defTZ, idemTTL: idemTTL,
		deliveryBase: deliveryBase, deliveryPerKm: deliveryPerKm,
	}
}

const idemScope = "orders:create"

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	claim, err := uc.idempotency.Claim(ctx, idempotencyrepo.ClaimIn{
		Scope: idemScope, Key: in.IdempotencyKey, RequestHash: in.RequestHash,
		ExpiresAt: uc.nower.Now().Add(uc.idemTTL),
	})
	if err != nil {
		return nil, err
	}
	if claim == idempotencyrepo.ClaimExists {
		rec, getErr := uc.idempotency.Get(ctx, idemScope, in.IdempotencyKey)
		if getErr != nil {
			return nil, getErr
		}
		if rec.RequestHash != in.RequestHash {
			return nil, usecase.ErrIdempotencyKeyReuse
		}
		if rec.ResponseStatus != nil && rec.ResponseBody != nil {
			return &Out{Replayed: true, ReplayBody: rec.ResponseBody, ReplayCode: *rec.ResponseStatus}, nil
		}
		return nil, usecase.ErrRequestInProgress
	}

	var created *orderrepo.Order
	if doErr := uc.trm.Do(ctx, func(ctx context.Context) error {
		var createErr error
		created, createErr = uc.createOrder(ctx, in)
		return createErr
	}); doErr != nil {
		_ = uc.idempotency.Delete(ctx, idemScope, in.IdempotencyKey)
		return nil, doErr
	}
	return &Out{Order: created}, nil
}

func (uc *UseCase) createOrder(ctx context.Context, in In) (*orderrepo.Order, error) {
	cart, err := uc.carts.GetActiveForUpdate(ctx, in.CustomerID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrCartNotFound
		}
		return nil, err
	}
	items, err := uc.carts.GetItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, usecase.ErrCartEmpty
	}

	venue, err := uc.venues.GetByID(ctx, cart.VenueID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrVenueNotFound
		}
		return nil, err
	}
	if venue.Status != "active" {
		return nil, usecase.ErrVenueNotFound
	}
	if !venue.AcceptsOrders {
		return nil, fmt.Errorf("%w: %v", usecase.ErrVenueClosed, errdetails.VenueClosedDetail("paused_by_venue"))
	}
	hours, _ := uc.venues.GetWorkingHours(ctx, venue.ID)
	isOpen, reason := shared.VenueIsOpen(venue, hours, uc.nower.Now(), uc.defTZ)
	if !isOpen {
		return nil, fmt.Errorf("%w: %v", usecase.ErrVenueClosed, errdetails.VenueClosedDetail(reason))
	}

	ft := orderstatus.FulfillmentType(in.FulfillmentType)
	if ft != orderstatus.FulfillmentDelivery && ft != orderstatus.FulfillmentPickup {
		return nil, usecase.ErrValidation
	}

	var distanceM int
	deliveryMin := 0
	deliveryFee := money.Minor(0)
	if ft == orderstatus.FulfillmentDelivery {
		var addr struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}
		if len(in.DeliveryAddress) == 0 || json.Unmarshal(in.DeliveryAddress, &addr) != nil {
			return nil, usecase.ErrValidation
		}
		distanceM = int(math.Round(geo.Haversine(venue.Lat, venue.Lon, addr.Lat, addr.Lon)))
		if distanceM > venue.DeliveryRadiusM {
			return nil, fmt.Errorf("%w: %v", usecase.ErrOutsideDeliveryZone,
				errdetails.OutsideDeliveryZoneDetail(distanceM, venue.DeliveryRadiusM))
		}
		deliveryFee = money.Minor(venue.DeliveryFeeMinor)
		distanceKm := float64(distanceM) / 1000.0
		deliveryMin = uc.deliveryBase + int(math.Ceil(distanceKm*uc.deliveryPerKm))
	}

	ids := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.MenuItemID)
	}
	menuItems, err := uc.catalog.GetItemsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	menuByID := make(map[uuid.UUID]catalogrepo.MenuItem, len(menuItems))
	for _, mi := range menuItems {
		menuByID[mi.ID] = mi
	}

	var itemsTotal money.Minor
	createItems := make([]orderrepo.CreateItemIn, 0, len(items))
	for _, it := range items {
		mi, ok := menuByID[it.MenuItemID]
		if !ok || !mi.IsActive {
			return nil, usecase.ErrMenuItemNotFound
		}
		if !mi.IsAvailable {
			return nil, fmt.Errorf("%w: %v", usecase.ErrMenuItemUnavailable, errdetails.UnavailableItemIDs(mi.ID))
		}
		optSnaps, optErr := shared.ResolveLineOptions(mi, it.SelectedOptions)
		if optErr != nil {
			return nil, usecase.ErrInvalidOptions
		}
		unit := shared.LineUnitPrice(mi, optSnaps)
		lineTotal := money.LineTotal(unit, it.Quantity)
		itemsTotal += lineTotal
		optsRaw, rawErr := shared.OptionsToRaw(optSnaps)
		if rawErr != nil {
			return nil, rawErr
		}
		mid := mi.ID
		createItems = append(createItems, orderrepo.CreateItemIn{
			MenuItemID: &mid, ItemExternalID: mi.ExternalID, NameSnapshot: mi.Name,
			UnitPriceMinor: int64(unit), Quantity: it.Quantity,
			OptionsSnapshot: optsRaw, LineTotalMinor: int64(lineTotal),
		})
	}

	if int64(itemsTotal) < venue.MinOrderMinor {
		return nil, fmt.Errorf("%w: %v", usecase.ErrBelowMinOrder,
			errdetails.BelowMinOrderDetail(venue.MinOrderMinor, int64(itemsTotal)))
	}

	total := itemsTotal + deliveryFee
	if in.ExpectedTotalMinor != nil && *in.ExpectedTotalMinor != int64(total) {
		return nil, fmt.Errorf("%w: %v", usecase.ErrPriceChanged,
			errdetails.PriceChangedDetail(*in.ExpectedTotalMinor, int64(total)))
	}

	paymentStatus := "not_required"
	if in.PaymentMethod == "card_online_mock" {
		paymentStatus = "paid"
	}

	prep := venue.AvgPrepMinutes
	eta := uc.nower.Now().Add(time.Duration(prep+deliveryMin) * time.Minute)
	payload, _ := json.Marshal(map[string]any{"order_id": cart.ID.String()})

	o, err := uc.orders.Create(ctx, orderrepo.CreateIn{
		CustomerID: in.CustomerID, VenueID: cart.VenueID, CartID: cart.ID,
		FulfillmentType: ft, DeliveryAddress: in.DeliveryAddress, ContactPhone: in.ContactPhone,
		ItemsTotalMinor: int64(itemsTotal), DeliveryFeeMinor: int64(deliveryFee),
		TotalMinor: int64(total), PaymentMethod: in.PaymentMethod, PaymentStatus: paymentStatus,
		Comment: in.Comment, PrepMinutes: prep, ETAAt: eta, Items: createItems,
		OutboxPayload: payload, CreateDelivery: ft == orderstatus.FulfillmentDelivery, DeliveryETAMin: deliveryMin,
	})
	if err != nil {
		return nil, err
	}
	return o, nil
}
