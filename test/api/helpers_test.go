//go:build e2e

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	customerID           = "11111111-1111-1111-1111-111111111111"
	venueID              = "11111111-1111-4111-8111-111111111101"
	burgerVenueID        = "11111111-1111-4111-8111-111111111103"
	menuItemID           = "33333333-3333-4333-8333-333333333301"
	burgerMenuItemID     = "33333333-3333-4333-8333-333333333401"
	sizeMedium           = "66666666-6666-4666-8666-666666666302"
	margheritaExternalID = "item-margherita"
	partnerAPIKey        = "ak_demopizz_xK9mN2pQ7vR4wL8hT3jF6sD1aB0cE5gY9uI2oP4"
	expectedTotal        = int64(137900) // 2×(49000+10000) + 19900 delivery
	burgerDeliveryTotal  = int64(59900)  // 45000 + 14900 delivery
)

func apiBase() string {
	if v := os.Getenv("API_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

func simulatorBase() string {
	if v := os.Getenv("SIMULATOR_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8081"
}

func apiReady(base string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/health/live", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

func requireAPI(t *testing.T) string {
	t.Helper()
	base := apiBase()
	if !apiReady(base) {
		t.Skip("API not available at " + base)
	}
	return base
}

func requireSimulator(t *testing.T) string {
	t.Helper()
	base := simulatorBase()
	if !apiReady(base) {
		t.Skip("simulator not available at " + base)
	}
	return base
}

func customerRequest(t *testing.T, method, url string, body any, headers map[string]string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	req.Header.Set("X-Customer-Id", customerID)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func simulatorRequest(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSONBody(t *testing.T, resp *http.Response, dest any) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(dest))
}

func readErrorCode(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	return payload.Error.Code
}

func addCartItem(t *testing.T, base, venue, item string, qty int, optionIDs []string) {
	t.Helper()
	if optionIDs == nil {
		optionIDs = []string{}
	}
	resp := customerRequest(t, http.MethodPost, base+"/api/v1/cart/items", map[string]any{
		"venue_id":     venue,
		"menu_item_id": item,
		"quantity":     qty,
		"option_ids":   optionIDs,
	}, nil)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func addDefaultCartItem(t *testing.T, base string) {
	addCartItem(t, base, venueID, menuItemID, 2, []string{sizeMedium})
}

func createOrderBody(total int64) map[string]any {
	return map[string]any{
		"fulfillment_type":     "delivery",
		"address":              map[string]any{"text": "Moscow, Tverskaya 1", "lat": 55.7558, "lon": 37.6176},
		"contact_phone":        "+79001234567",
		"payment_method":       "cash",
		"expected_total_minor": total,
	}
}

func createOrder(t *testing.T, base string, total int64, idemKey string) string {
	t.Helper()
	addDefaultCartItem(t, base)

	resp := customerRequest(t, http.MethodPost, base+"/api/v1/orders", createOrderBody(total), map[string]string{
		"Idempotency-Key": idemKey,
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		Order struct {
			ID string `json:"id"`
		} `json:"order"`
	}
	decodeJSONBody(t, resp, &created)
	require.NotEmpty(t, created.Order.ID)
	return created.Order.ID
}

func setSimulatorMode(t *testing.T, simBase, mode string) {
	t.Helper()
	resp := simulatorRequest(t, http.MethodPost, simBase+"/control/mode", map[string]any{"mode": mode})
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func setStoplist(t *testing.T, simBase, externalID string, available bool) {
	t.Helper()
	resp := simulatorRequest(t, http.MethodPost, simBase+"/control/stoplist", map[string]any{
		"external_id":   externalID,
		"is_available": available,
	})
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func pollOrderStatus(t *testing.T, base, orderID string, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp := customerRequest(t, http.MethodGet, base+"/api/v1/orders/"+orderID, nil, nil)
		var payload struct {
			Order struct {
				Status       string  `json:"status"`
				CancelReason *string `json:"cancel_reason"`
			} `json:"order"`
		}
		decodeJSONBody(t, resp, &payload)
		if payload.Order.Status == want {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("order %s did not reach status %q within %s", orderID, want, timeout)
}

func pollOrderField(t *testing.T, base, orderID string, timeout time.Duration, check func(status string, cancelReason *string) bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp := customerRequest(t, http.MethodGet, base+"/api/v1/orders/"+orderID, nil, nil)
		var payload struct {
			Order struct {
				Status       string  `json:"status"`
				CancelReason *string `json:"cancel_reason"`
			} `json:"order"`
		}
		decodeJSONBody(t, resp, &payload)
		if check(payload.Order.Status, payload.Order.CancelReason) {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("order %s condition not met within %s", orderID, timeout)
}

func clearCart(t *testing.T, base string) {
	t.Helper()
	resp := customerRequest(t, http.MethodDelete, base+"/api/v1/cart", nil, nil)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return
	}
}

func uniqueKey(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func resetSimulatorDefaults(t *testing.T) {
	t.Helper()
	resetSimulatorState()
}

func resetSimulatorState() {
	sim := simulatorBase()
	if !apiReady(sim) {
		return
	}
	body, _ := json.Marshal(map[string]any{"mode": "auto_accept"})
	req, _ := http.NewRequest(http.MethodPost, sim+"/control/mode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	body, _ = json.Marshal(map[string]any{"external_id": margheritaExternalID, "is_available": true})
	req, _ = http.NewRequest(http.MethodPost, sim+"/control/stoplist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	cleanupStalePlacedOrders()
}

func cleanupStalePlacedOrders() {
	base := apiBase()
	if !apiReady(base) {
		return
	}
	for page := 0; page < 20; page++ {
		url := base + "/partner/v1/orders?status=placed&limit=100"
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "Bearer "+partnerAPIKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		var payload struct {
			Items []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		_ = resp.Body.Close()
		if decodeErr != nil || resp.StatusCode != http.StatusOK {
			return
		}
		rejected := 0
		for _, item := range payload.Items {
			if item.Status != "placed" {
				continue
			}
			body, _ := json.Marshal(map[string]any{
				"reason":      "e2e cleanup",
				"reason_code": "out_of_stock",
			})
			rejectReq, err := http.NewRequest(http.MethodPost, base+"/partner/v1/orders/"+item.ID+"/reject", bytes.NewReader(body))
			if err != nil {
				continue
			}
			rejectReq.Header.Set("Authorization", "Bearer "+partnerAPIKey)
			rejectReq.Header.Set("Content-Type", "application/json")
			rejectResp, err := http.DefaultClient.Do(rejectReq)
			if err == nil {
				_ = rejectResp.Body.Close()
				rejected++
			}
		}
		if rejected == 0 {
			return
		}
	}
}

func waitForNoPlacedOrders(timeout time.Duration) {
	base := apiBase()
	if !apiReady(base) {
		return
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, base+"/partner/v1/orders?status=placed&limit=1", nil)
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "Bearer "+partnerAPIKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		var payload struct {
			Items []any `json:"items"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		_ = resp.Body.Close()
		if decodeErr != nil || resp.StatusCode != http.StatusOK {
			return
		}
		if len(payload.Items) == 0 {
			return
		}
		cleanupStalePlacedOrders()
		time.Sleep(500 * time.Millisecond)
	}
}
