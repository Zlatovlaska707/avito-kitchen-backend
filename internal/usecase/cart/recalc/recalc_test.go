package recalc_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/recalc"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

func sampleVenue() *venuerepo.Venue {
	return &venuerepo.Venue{
		ID: uuid.New(), Name: "Pizza Mario", MinOrderMinor: 50000, DeliveryFeeMinor: 19900,
	}
}

func sampleMenuItem(id uuid.UUID, price int64, available bool) catalogrepo.MenuItem {
	return catalogrepo.MenuItem{
		ID: id, ExternalID: "item-1", Name: "Margherita", PriceMinor: price,
		IsActive: true, IsAvailable: available,
	}
}

func TestRun_NilCart(t *testing.T) {
	result := recalc.Run(nil, nil, nil, sampleVenue(), true, "")
	assert.Nil(t, result.Cart)
	assert.Empty(t, result.Lines)
}

func TestRun_VenueClosedIssue(t *testing.T) {
	cart := &cartrepo.Cart{ID: uuid.New(), VenueID: uuid.New()}
	venue := sampleVenue()
	venue.MinOrderMinor = 0
	result := recalc.Run(cart, nil, nil, venue, false, "schedule")
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "venue_closed", result.Issues[0].Code)
	assert.Equal(t, "schedule", result.Issues[0].Details["reason"])
}

func TestRun_ItemRemoved(t *testing.T) {
	cart := &cartrepo.Cart{ID: uuid.New(), VenueID: uuid.New()}
	itemID := uuid.New()
	items := []cartrepo.Item{{
		ID: itemID, MenuItemID: itemID, Quantity: 1, SelectedOptions: json.RawMessage(`[]`),
	}}
	venue := sampleVenue()
	venue.MinOrderMinor = 0
	result := recalc.Run(cart, items, nil, venue, true, "")
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "item_removed", result.Issues[0].Code)
	assert.Empty(t, result.Lines)
}

func TestRun_ItemUnavailable(t *testing.T) {
	cart := &cartrepo.Cart{ID: uuid.New(), VenueID: uuid.New()}
	itemID := uuid.New()
	items := []cartrepo.Item{{
		ID: uuid.New(), MenuItemID: itemID, Quantity: 2, SelectedOptions: json.RawMessage(`[]`),
	}}
	menu := []catalogrepo.MenuItem{sampleMenuItem(itemID, 49000, false)}

	result := recalc.Run(cart, items, menu, sampleVenue(), true, "")
	require.NotEmpty(t, result.Lines)
	assert.Equal(t, int64(98000), result.ItemsTotalMinor)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "item_unavailable", result.Issues[0].Code)
}

func TestRun_BelowMinOrder(t *testing.T) {
	cart := &cartrepo.Cart{ID: uuid.New(), VenueID: uuid.New()}
	itemID := uuid.New()
	items := []cartrepo.Item{{
		ID: uuid.New(), MenuItemID: itemID, Quantity: 1, SelectedOptions: json.RawMessage(`[]`),
	}}
	menu := []catalogrepo.MenuItem{sampleMenuItem(itemID, 10000, true)}
	venue := sampleVenue()
	venue.MinOrderMinor = 50000

	result := recalc.Run(cart, items, menu, venue, true, "")
	codes := issueCodes(result.Issues)
	assert.Contains(t, codes, "below_min_order")
	assert.Equal(t, int64(10000), result.ItemsTotalMinor)
	assert.Equal(t, int64(29900), result.TotalMinor) // items + delivery fee
}

func TestRun_HappyPathWithDeliveryFee(t *testing.T) {
	cart := &cartrepo.Cart{ID: uuid.New(), VenueID: uuid.New()}
	itemID := uuid.New()
	items := []cartrepo.Item{{
		ID: uuid.New(), MenuItemID: itemID, Quantity: 2, SelectedOptions: json.RawMessage(`[]`),
	}}
	menu := []catalogrepo.MenuItem{sampleMenuItem(itemID, 49000, true)}
	venue := sampleVenue()

	result := recalc.Run(cart, items, menu, venue, true, "")
	assert.Empty(t, result.Issues)
	assert.Equal(t, int64(98000), result.ItemsTotalMinor)
	assert.Equal(t, int64(19900), result.DeliveryFeeMinor)
	assert.Equal(t, int64(117900), result.TotalMinor)
	assert.Equal(t, "Pizza Mario", result.VenueName)
}

func TestRun_InactiveMenuItem(t *testing.T) {
	cart := &cartrepo.Cart{ID: uuid.New(), VenueID: uuid.New()}
	itemID := uuid.New()
	items := []cartrepo.Item{{
		ID: uuid.New(), MenuItemID: itemID, Quantity: 1,
		SelectedOptions: json.RawMessage(`[]`), CreatedAt: time.Now(),
	}}
	mi := sampleMenuItem(itemID, 49000, true)
	mi.IsActive = false
	result := recalc.Run(cart, items, []catalogrepo.MenuItem{mi}, sampleVenue(), true, "")
	assert.Equal(t, "item_removed", result.Issues[0].Code)
}

func issueCodes(issues []recalc.Issue) []string {
	out := make([]string, len(issues))
	for i, iss := range issues {
		out[i] = iss.Code
	}
	return out
}
