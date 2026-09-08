package list_for_admin_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_admin"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func makeOrders(n int, venueID uuid.UUID) []orderrepo.Order {
	orders := make([]orderrepo.Order, n)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := range orders {
		orders[i] = orderrepo.Order{
			ID:        uuid.New(),
			VenueID:   venueID,
			Status:    orderstatus.StatusPlaced,
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
		}
	}
	return orders
}

func TestRun_HappyPath(t *testing.T) {
	venueID := uuid.New()
	orders := makeOrders(2, venueID)
	repo := &uctest.OrderRepo{ListOrders: orders}
	uc := list_for_admin.New(repo)

	out, err := uc.Run(context.Background(), list_for_admin.In{
		VenueID: &venueID,
		Limit:   10,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Len(t, out.Orders, 2)
	assert.False(t, out.HasNext)
	require.NotNil(t, out.LastTS)
	require.NotNil(t, out.LastID)
}

func TestRun_DefaultLimit(t *testing.T) {
	venueID := uuid.New()
	orders := makeOrders(21, venueID)
	repo := &uctest.OrderRepo{ListOrders: orders}
	uc := list_for_admin.New(repo)

	out, err := uc.Run(context.Background(), list_for_admin.In{VenueID: &venueID})
	require.NoError(t, err)
	assert.Len(t, out.Orders, 20)
	assert.True(t, out.HasNext)
}

func TestRun_HasNextPagination(t *testing.T) {
	venueID := uuid.New()
	orders := makeOrders(3, venueID)
	repo := &uctest.OrderRepo{ListOrders: orders}
	uc := list_for_admin.New(repo)

	out, err := uc.Run(context.Background(), list_for_admin.In{
		VenueID: &venueID,
		Limit:   2,
	})
	require.NoError(t, err)
	assert.Len(t, out.Orders, 2)
	assert.True(t, out.HasNext)
}

func TestRun_EmptyResult(t *testing.T) {
	repo := &uctest.OrderRepo{ListOrders: nil}
	uc := list_for_admin.New(repo)

	out, err := uc.Run(context.Background(), list_for_admin.In{Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, out.Orders)
	assert.False(t, out.HasNext)
	assert.Nil(t, out.LastTS)
	assert.Nil(t, out.LastID)
}

func TestRun_WithStatusFilter(t *testing.T) {
	venueID := uuid.New()
	status := string(orderstatus.StatusAccepted)
	repo := &uctest.OrderRepo{
		ListOrders: []orderrepo.Order{{ID: uuid.New(), VenueID: venueID, Status: orderstatus.StatusAccepted}},
	}
	uc := list_for_admin.New(repo)

	out, err := uc.Run(context.Background(), list_for_admin.In{
		VenueID: &venueID,
		Status:  &status,
		Limit:   10,
	})
	require.NoError(t, err)
	assert.Len(t, out.Orders, 1)
}

func TestRun_ListError(t *testing.T) {
	repo := &uctest.OrderRepo{ListErr: assert.AnError}
	uc := list_for_admin.New(repo)

	_, err := uc.Run(context.Background(), list_for_admin.In{Limit: 10})
	assert.ErrorIs(t, err, assert.AnError)
}
