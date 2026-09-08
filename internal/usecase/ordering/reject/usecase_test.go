package reject_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/reject"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_EmptyReason(t *testing.T) {
	uc := reject.New(&uctest.OrderRepo{}, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), reject.In{OrderID: uuid.New(), VenueID: uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrValidation)
}

func TestRun_HappyPath(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()
	venueID := uuid.New()

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: now,
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusRejected},
	}
	uc := reject.New(orders, uctest.FixedNower{T: now}, time.Minute, 10*time.Minute)

	out, err := uc.Run(context.Background(), reject.In{
		OrderID: orderID, VenueID: venueID, Reason: "out of stock",
	})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.Equal(t, orderstatus.StatusRejected, out.Order.Status)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByIDErr: pgerr.ErrNotFound}
	uc := reject.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), reject.In{
		OrderID: uuid.New(), VenueID: uuid.New(), Reason: "busy",
	})
	assert.ErrorIs(t, err, usecase.ErrOrderNotFound)
}

func TestRun_InvalidTransition(t *testing.T) {
	orderID := uuid.New()
	venueID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusCompleted,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	uc := reject.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), reject.In{
		OrderID: orderID, VenueID: venueID, Reason: "too late",
	})
	assert.ErrorIs(t, err, usecase.ErrInvalidTransition)
}

func TestRun_TransitionRace(t *testing.T) {
	orderID := uuid.New()
	venueID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
		TransitionErr: pgerr.ErrNotFound,
	}
	uc := reject.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), reject.In{
		OrderID: orderID, VenueID: venueID, Reason: "closed",
	})
	assert.ErrorIs(t, err, usecase.ErrInvalidTransition)
}
