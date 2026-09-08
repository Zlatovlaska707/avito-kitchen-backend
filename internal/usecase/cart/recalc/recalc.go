package recalc

import (
	"github.com/zlatovlaska707/avito-kitchen/internal/domain/money"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type SelectedOption struct {
	ID              string
	ExternalID      string
	Name            string
	PriceDeltaMinor int64
}

type Line struct {
	ID              string
	MenuItemID      string
	Name            string
	Quantity        int
	UnitPriceMinor  int64
	LineTotalMinor  int64
	SelectedOptions []SelectedOption
}

type Issue struct {
	Code    string
	Message string
	Details map[string]any
}

type Result struct {
	Cart             *cartrepo.Cart
	VenueName        string
	Lines            []Line
	ItemsTotalMinor  int64
	DeliveryFeeMinor int64
	TotalMinor       int64
	Issues           []Issue
}

func Run(cart *cartrepo.Cart, items []cartrepo.Item, menuItems []catalogrepo.MenuItem, venue *venuerepo.Venue, venueOpen bool, closedReason string) Result {
	if cart == nil {
		return Result{}
	}

	menuByID := make(map[string]catalogrepo.MenuItem, len(menuItems))
	for _, m := range menuItems {
		menuByID[m.ID.String()] = m
	}

	var issues []Issue
	if !venueOpen {
		issues = append(issues, Issue{
			Code: "venue_closed", Message: "Venue is closed",
			Details: map[string]any{"reason": closedReason},
		})
	}

	var lines []Line
	var itemsTotal money.Minor
	for _, it := range items {
		mi, ok := menuByID[it.MenuItemID.String()]
		if !ok || !mi.IsActive {
			issues = append(issues, Issue{
				Code: "item_removed", Message: "Item removed from menu",
				Details: map[string]any{"cart_item_id": it.ID.String()},
			})
			continue
		}
		if !mi.IsAvailable {
			issues = append(issues, Issue{
				Code: "item_unavailable", Message: "Item unavailable",
				Details: map[string]any{"cart_item_id": it.ID.String()},
			})
		}
		optSnaps, err := shared.ResolveLineOptions(mi, it.SelectedOptions)
		if err != nil {
			issues = append(issues, Issue{
				Code: "option_unavailable", Message: "Selected option unavailable",
				Details: map[string]any{"cart_item_id": it.ID.String()},
			})
			continue
		}
		unit := shared.LineUnitPrice(mi, optSnaps)
		selected := make([]SelectedOption, 0, len(optSnaps))
		for _, o := range optSnaps {
			selected = append(selected, SelectedOption{
				ID: o.OptionID.String(), ExternalID: o.ExternalID, Name: o.Name, PriceDeltaMinor: o.PriceDeltaMinor,
			})
		}
		line := Line{
			ID: it.ID.String(), MenuItemID: it.MenuItemID.String(),
			Name: mi.Name, Quantity: it.Quantity,
			UnitPriceMinor: int64(unit), LineTotalMinor: int64(money.LineTotal(unit, it.Quantity)),
			SelectedOptions: selected,
		}
		lines = append(lines, line)
		itemsTotal += money.Minor(line.LineTotalMinor)
	}

	deliveryFee := money.Minor(0)
	if venue != nil {
		deliveryFee = money.Minor(venue.DeliveryFeeMinor)
	}
	total := itemsTotal + deliveryFee

	if venue != nil && int64(itemsTotal) < venue.MinOrderMinor {
		issues = append(issues, Issue{
			Code: "below_min_order", Message: "Below minimum order",
			Details: map[string]any{
				"min_order_minor": venue.MinOrderMinor, "items_total_minor": int64(itemsTotal),
			},
		})
	}

	venueName := ""
	if venue != nil {
		venueName = venue.Name
	}

	return Result{
		Cart: cart, VenueName: venueName, Lines: lines,
		ItemsTotalMinor: int64(itemsTotal), DeliveryFeeMinor: int64(deliveryFee),
		TotalMinor: int64(total), Issues: issues,
	}
}
