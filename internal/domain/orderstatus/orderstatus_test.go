package orderstatus_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
)

func tc(now time.Time, acceptedAt *time.Time, grace, acceptTimeout time.Duration, ft orderstatus.FulfillmentType, reason string, timeoutHit bool) orderstatus.TransitionContext {
	return orderstatus.TransitionContext{
		Now: now, AcceptedAt: acceptedAt, CancelGrace: grace, AcceptTimeout: acceptTimeout,
		PlacedAt: now.Add(-acceptTimeout), FulfillmentType: ft, Reason: reason,
		SystemTimeoutHit: timeoutHit,
	}
}

func TestCanTransition_PlacedToAccepted(t *testing.T) {
	now := time.Now()
	err := orderstatus.CanTransition(orderstatus.StatusPlaced, orderstatus.StatusAccepted, orderstatus.ActorVenue, tc(now, nil, 0, time.Minute, orderstatus.FulfillmentDelivery, "", false))
	assert.NoError(t, err)
}

func TestCanTransition_PlacedToAccepted_CustomerDenied(t *testing.T) {
	err := orderstatus.CanTransition(orderstatus.StatusPlaced, orderstatus.StatusAccepted, orderstatus.ActorCustomer, tc(time.Now(), nil, 0, time.Minute, orderstatus.FulfillmentDelivery, "", false))
	assert.ErrorIs(t, err, orderstatus.ErrInvalidTransition)
}

func TestCanTransition_PlacedToRejected_RequiresReason(t *testing.T) {
	now := time.Now()
	err := orderstatus.CanTransition(orderstatus.StatusPlaced, orderstatus.StatusRejected, orderstatus.ActorVenue, tc(now, nil, 0, time.Minute, orderstatus.FulfillmentDelivery, "", false))
	assert.ErrorIs(t, err, orderstatus.ErrInvalidTransition)

	err = orderstatus.CanTransition(orderstatus.StatusPlaced, orderstatus.StatusRejected, orderstatus.ActorVenue, tc(now, nil, 0, time.Minute, orderstatus.FulfillmentDelivery, "busy", false))
	assert.NoError(t, err)
}

func TestCanTransition_PlacedToCancelled(t *testing.T) {
	now := time.Now()
	ctx := tc(now, nil, 0, 5*time.Minute, orderstatus.FulfillmentDelivery, "", false)
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusPlaced, orderstatus.StatusCancelled, orderstatus.ActorCustomer, ctx))
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusPlaced, orderstatus.StatusCancelled, orderstatus.ActorAdmin, ctx))
}

func TestCanTransition_PlacedToCancelled_SystemTimeout(t *testing.T) {
	now := time.Now()
	ctx := tc(now, nil, 0, 5*time.Minute, orderstatus.FulfillmentDelivery, "", true)
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusPlaced, orderstatus.StatusCancelled, orderstatus.ActorSystem, ctx))
}

func TestCanTransition_AcceptedToReady(t *testing.T) {
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusAccepted, orderstatus.StatusReady, orderstatus.ActorVenue, tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentDelivery, "", false)))
}

func TestCanTransition_AcceptedCancel_GracePeriod(t *testing.T) {
	now := time.Now()
	accepted := now.Add(-2 * time.Minute)
	grace := 5 * time.Minute
	ctx := tc(now, &accepted, grace, 0, orderstatus.FulfillmentDelivery, "", false)

	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusAccepted, orderstatus.StatusCancelled, orderstatus.ActorCustomer, ctx))

	afterGrace := now.Add(4 * time.Minute)
	ctxAfter := tc(afterGrace, &accepted, grace, 0, orderstatus.FulfillmentDelivery, "", false)
	err := orderstatus.CanTransition(orderstatus.StatusAccepted, orderstatus.StatusCancelled, orderstatus.ActorCustomer, ctxAfter)
	assert.ErrorIs(t, err, orderstatus.ErrNotCancellable)
}

func TestCanTransition_AcceptedCancel_ExactlyAtGrace(t *testing.T) {
	now := time.Now()
	accepted := now.Add(-5 * time.Minute)
	grace := 5 * time.Minute
	ctx := tc(now, &accepted, grace, 0, orderstatus.FulfillmentDelivery, "", false)
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusAccepted, orderstatus.StatusCancelled, orderstatus.ActorCustomer, ctx))
}

func TestCanTransition_ReadyToInDelivery(t *testing.T) {
	ctx := tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentDelivery, "", false)
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusReady, orderstatus.StatusInDelivery, orderstatus.ActorVenue, ctx))
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusReady, orderstatus.StatusInDelivery, orderstatus.ActorSystem, ctx))

	ctxPickup := tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentPickup, "", false)
	err := orderstatus.CanTransition(orderstatus.StatusReady, orderstatus.StatusInDelivery, orderstatus.ActorVenue, ctxPickup)
	assert.ErrorIs(t, err, orderstatus.ErrInvalidTransition)
}

func TestCanTransition_ReadyToCompleted_Pickup(t *testing.T) {
	ctx := tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentPickup, "", false)
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusReady, orderstatus.StatusCompleted, orderstatus.ActorVenue, ctx))
}

func TestCanTransition_InDeliveryToCompleted(t *testing.T) {
	ctx := tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentDelivery, "", false)
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusInDelivery, orderstatus.StatusCompleted, orderstatus.ActorSystem, ctx))
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusInDelivery, orderstatus.StatusCompleted, orderstatus.ActorCourier, ctx))
}

func TestCanTransition_InDeliveryAdminCancel(t *testing.T) {
	ctx := tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentDelivery, "", false)
	assert.NoError(t, orderstatus.CanTransition(orderstatus.StatusInDelivery, orderstatus.StatusCancelled, orderstatus.ActorAdmin, ctx))
}

func TestCanTransition_TerminalStates(t *testing.T) {
	terminals := []orderstatus.Status{orderstatus.StatusCompleted, orderstatus.StatusCancelled, orderstatus.StatusRejected}
	ctx := tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentDelivery, "", false)
	for _, from := range terminals {
		err := orderstatus.CanTransition(from, orderstatus.StatusAccepted, orderstatus.ActorVenue, ctx)
		require.ErrorIs(t, err, orderstatus.ErrInvalidTransition)
	}
}

func TestCanTransition_InvalidPairs(t *testing.T) {
	ctx := tc(time.Now(), nil, 0, 0, orderstatus.FulfillmentDelivery, "", false)
	pairs := []struct {
		from, to orderstatus.Status
		actor    orderstatus.ActorType
	}{
		{orderstatus.StatusPlaced, orderstatus.StatusReady, orderstatus.ActorVenue},
		{orderstatus.StatusAccepted, orderstatus.StatusCompleted, orderstatus.ActorVenue},
		{orderstatus.StatusReady, orderstatus.StatusAccepted, orderstatus.ActorVenue},
	}
	for _, p := range pairs {
		err := orderstatus.CanTransition(p.from, p.to, p.actor, ctx)
		assert.ErrorIs(t, err, orderstatus.ErrInvalidTransition, "%s -> %s", p.from, p.to)
	}
}
