package handover_test

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
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/handover"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_HappyPathPickup(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()
	venueID := uuid.New()
	acceptedAt := now.Add(-20 * time.Minute)

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusReady,
			FulfillmentType: orderstatus.FulfillmentPickup, PlacedAt: now.Add(-30 * time.Minute),
			AcceptedAt: &acceptedAt,
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusCompleted},
	}
	uc := handover.New(orders, uctest.FixedNower{T: now}, time.Minute, 10*time.Minute)

	out, err := uc.Run(context.Background(), handover.In{OrderID: orderID, VenueID: venueID})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.Equal(t, orderstatus.StatusCompleted, out.Order.Status)
}

func TestRun_HappyPathDelivery(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()
	venueID := uuid.New()
	acceptedAt := now.Add(-20 * time.Minute)

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusReady,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: now.Add(-30 * time.Minute),
			AcceptedAt: &acceptedAt,
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusInDelivery},
	}
	uc := handover.New(orders, uctest.FixedNower{T: now}, time.Minute, 10*time.Minute)

	out, err := uc.Run(context.Background(), handover.In{OrderID: orderID, VenueID: venueID})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.Equal(t, orderstatus.StatusInDelivery, out.Order.Status)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByIDErr: pgerr.ErrNotFound}
	uc := handover.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), handover.In{OrderID: uuid.New(), VenueID: uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrOrderNotFound)
}

func TestRun_InvalidTransition(t *testing.T) {
	orderID := uuid.New()
	venueID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusAccepted,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	uc := handover.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), handover.In{OrderID: orderID, VenueID: venueID})
	assert.ErrorIs(t, err, usecase.ErrInvalidTransition)
}

func TestRun_TransitionRace(t *testing.T) {
	orderID := uuid.New()
	venueID := uuid.New()
	acceptedAt := time.Now().Add(-time.Minute)
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusReady,
			FulfillmentType: orderstatus.FulfillmentPickup, PlacedAt: time.Now(),
			AcceptedAt: &acceptedAt,
		},
		TransitionErr: pgerr.ErrNotFound,
	}
	uc := handover.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), handover.In{OrderID: orderID, VenueID: venueID})
	assert.ErrorIs(t, err, usecase.ErrInvalidTransition)
}
