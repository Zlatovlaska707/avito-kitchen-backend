package accept_test

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
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/accept"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_HappyPath(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()
	venueID := uuid.New()
	prep := 25

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: venueID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: now.Add(-time.Minute),
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusAccepted},
	}
	uc := accept.New(orders, uctest.FixedNower{T: now}, time.Minute, 10*time.Minute)

	out, err := uc.Run(context.Background(), accept.In{OrderID: orderID, VenueID: venueID, PrepMinutes: &prep})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.Equal(t, orderstatus.StatusAccepted, out.Order.Status)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByIDErr: pgerr.ErrNotFound}
	uc := accept.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), accept.In{OrderID: uuid.New(), VenueID: uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrOrderNotFound)
}

func TestRun_WrongVenue(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, VenueID: uuid.New(), Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	uc := accept.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), accept.In{OrderID: orderID, VenueID: uuid.New()})
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
	uc := accept.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), accept.In{OrderID: orderID, VenueID: venueID})
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
	uc := accept.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), accept.In{OrderID: orderID, VenueID: venueID})
	assert.ErrorIs(t, err, usecase.ErrInvalidTransition)
}
