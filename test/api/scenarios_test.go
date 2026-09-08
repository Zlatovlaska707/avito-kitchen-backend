//go:build e2e

package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoReject(t *testing.T) {
	base := requireAPI(t)
	sim := requireSimulator(t)

	setSimulatorMode(t, sim, "auto_reject")
	t.Cleanup(func() { setSimulatorMode(t, sim, "auto_accept") })

	cleanupStalePlacedOrders()
	waitForNoPlacedOrders(10 * time.Second)
	clearCart(t, base)
	addDefaultCartItem(t, base)
	resp := customerRequest(t, http.MethodPost, base+"/api/v1/orders", createOrderBody(expectedTotal), map[string]string{
		"Idempotency-Key": uniqueKey("e2e-reject"),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		Order struct {
			ID string `json:"id"`
		} `json:"order"`
	}
	decodeJSONBody(t, resp, &created)

	pollOrderField(t, base, created.Order.ID, 60*time.Second, func(status string, cancelReason *string) bool {
		return status == "rejected" && cancelReason != nil && *cancelReason != ""
	})
}

func TestStoplistBlocksOrder(t *testing.T) {
	base := requireAPI(t)
	sim := requireSimulator(t)

	clearCart(t, base)
	addDefaultCartItem(t, base)

	setStoplist(t, sim, margheritaExternalID, false)
	t.Cleanup(func() { setStoplist(t, sim, margheritaExternalID, true) })

	resp := customerRequest(t, http.MethodPost, base+"/api/v1/orders", createOrderBody(expectedTotal), map[string]string{
		"Idempotency-Key": uniqueKey("e2e-stoplist-block"),
	})
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	assert.Equal(t, "item_unavailable", readErrorCode(t, resp))

	setStoplist(t, sim, margheritaExternalID, true)
	resp = customerRequest(t, http.MethodPost, base+"/api/v1/orders", createOrderBody(expectedTotal), map[string]string{
		"Idempotency-Key": uniqueKey("e2e-stoplist-ok"),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestVenueClosedSchedule(t *testing.T) {
	base := requireAPI(t)

	resp := customerRequest(t, http.MethodGet, base+"/api/v1/venues?city_id=1&only_open=true", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list struct {
		Items []struct {
			Slug string `json:"slug"`
		} `json:"items"`
	}
	require.NoError(t, decodeListBody(resp, &list))
	for _, v := range list.Items {
		assert.NotEqual(t, "burger-closed", v.Slug)
	}

	resp = customerRequest(t, http.MethodGet, base+"/api/v1/venues/burger-closed", nil, nil)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var venue struct {
		IsOpenNow    bool    `json:"is_open_now"`
		ClosedReason *string `json:"closed_reason"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&venue))
	assert.False(t, venue.IsOpenNow)
	require.NotNil(t, venue.ClosedReason)
	assert.Equal(t, "schedule", *venue.ClosedReason)

	clearCart(t, base)
	addCartItem(t, base, burgerVenueID, burgerMenuItemID, 1, nil)
	resp = customerRequest(t, http.MethodPost, base+"/api/v1/orders", createOrderBody(burgerDeliveryTotal), map[string]string{
		"Idempotency-Key": uniqueKey("e2e-venue-closed"),
	})
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	assert.Equal(t, "venue_closed", readErrorCode(t, resp))

	t.Cleanup(func() { clearCart(t, base) })
}

func TestAcceptTimeoutSilent(t *testing.T) {
	base := requireAPI(t)
	sim := requireSimulator(t)

	setSimulatorMode(t, sim, "silent")
	t.Cleanup(func() { setSimulatorMode(t, sim, "auto_accept") })

	clearCart(t, base)
	addDefaultCartItem(t, base)
	resp := customerRequest(t, http.MethodPost, base+"/api/v1/orders", createOrderBody(expectedTotal), map[string]string{
		"Idempotency-Key": uniqueKey("e2e-silent"),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		Order struct {
			ID string `json:"id"`
		} `json:"order"`
	}
	decodeJSONBody(t, resp, &created)

	pollOrderField(t, base, created.Order.ID, 90*time.Second, func(status string, cancelReason *string) bool {
		return status == "cancelled" && cancelReason != nil && *cancelReason == "venue_no_response"
	})
}

func TestBrokenWebhookPollingFallback(t *testing.T) {
	base := requireAPI(t)
	sim := requireSimulator(t)

	setSimulatorMode(t, sim, "broken_webhook")
	t.Cleanup(func() { setSimulatorMode(t, sim, "auto_accept") })

	clearCart(t, base)
	orderID := createOrder(t, base, expectedTotal, uniqueKey("e2e-broken-webhook"))
	pollOrderStatus(t, base, orderID, "completed", 3*time.Minute)

	resp := simulatorRequest(t, http.MethodGet, sim+"/orders", nil)
	var queue struct {
		OrderIDs []string `json:"order_ids"`
	}
	decodeJSONBody(t, resp, &queue)
	assert.Contains(t, queue.OrderIDs, orderID)
}

func decodeListBody(resp *http.Response, dest any) error {
	defer func() { _ = resp.Body.Close() }()
	return json.NewDecoder(resp.Body).Decode(dest)
}
