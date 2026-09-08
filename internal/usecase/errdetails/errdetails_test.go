package errdetails_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/errdetails"
)

func TestConstructors(t *testing.T) {
	id := uuid.New()
	venueID := uuid.New()

	assert.Equal(t, []errdetails.FieldError{{Field: "f", Message: "m"}}, errdetails.ValidationFields(errdetails.FieldError{Field: "f", Message: "m"}).Fields)
	assert.Equal(t, []uuid.UUID{id}, errdetails.UnavailableItemIDs(id).ItemIDs)
	assert.Equal(t, venueID, errdetails.CartVenueMismatchDetail(venueID).CurrentVenueID)
	assert.Equal(t, int64(500), errdetails.BelowMinOrderDetail(500, 100).MinOrderMinor)
	assert.Equal(t, int64(100), errdetails.BelowMinOrderDetail(500, 100).ItemsTotalMinor)
	assert.Equal(t, 1500, errdetails.OutsideDeliveryZoneDetail(1500, 1000).DistanceM)
	assert.Equal(t, int64(200), errdetails.PriceChangedDetail(100, 200).ActualTotalMinor)
	assert.Equal(t, orderstatus.StatusReady, errdetails.InvalidTransitionDetail(orderstatus.StatusAccepted, orderstatus.StatusReady).RequestedStatus)
	assert.Equal(t, []string{"a"}, errdetails.DuplicateExternalIDs("a").ExternalIDs)
	assert.Equal(t, "schedule", errdetails.VenueClosedDetail("schedule").Reason)
}
