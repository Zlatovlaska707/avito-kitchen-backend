//go:build integration

package contract_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/app"
	"github.com/zlatovlaska707/avito-kitchen/internal/config"
	openapispec "github.com/zlatovlaska707/avito-kitchen/internal/openapi"
	"github.com/zlatovlaska707/avito-kitchen/test/suite"
)

const (
	adminToken     = "local-admin-token-change-me"
	partnerAPIKey  = "ak_demopizz_xK9mN2pQ7vR4wL8hT3jF6sD1aB0cE5gY9uI2oP4"
	customerID     = "11111111-1111-1111-1111-111111111111"
	venueSlug      = "pizza-mario"
	menuItemID     = "33333333-3333-4333-8333-333333333301"
	sizeMedium     = "66666666-6666-4666-8666-666666666302"
)

var domainErrorCodes = []string{
	"unauthorized",
	"forbidden",
	"validation_failed",
	"invalid_customer_id",
	"venue_not_found",
	"venue_slug_taken",
	"venue_closed",
	"menu_item_not_found",
	"item_unavailable",
	"invalid_options",
	"duplicate_external_id",
	"menu_sync_in_progress",
	"cart_not_found",
	"cart_empty",
	"cart_venue_mismatch",
	"cart_item_limit_exceeded",
	"quantity_limit_exceeded",
	"below_min_order",
	"outside_delivery_zone",
	"price_changed",
	"order_not_found",
	"invalid_transition",
	"not_cancellable",
	"idempotency_key_reuse",
	"request_in_progress",
	"webhook_url_forbidden",
	"webhook_already_exists",
	"payload_too_large",
	"rate_limited",
	"service_unavailable",
	"internal_error",
}

func TestOpenAPISpecCompleteness(t *testing.T) {
	doc, err := openapispec.Load()
	require.NoError(t, err)

	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			require.NotEmptyf(t, op.OperationID, "%s %s missing operationId", method, path)
			require.NotEmptyf(t, op.Tags, "%s %s missing tags", method, path)
			require.NotEmptyf(t, op.Summary, "%s %s missing summary", method, path)
			require.Truef(t, has4xxResponse(op.Responses), "%s %s missing 4xx response", method, path)
		}
	}
}

func TestErrorCodesInOpenAPISpec(t *testing.T) {
	specPath := filepath.Join(openapispec.ProjectRoot(), "api", "openapi.yaml")
	raw, err := os.ReadFile(specPath)
	require.NoError(t, err)
	content := string(raw)
	for _, code := range domainErrorCodes {
		require.Containsf(t, content, code, "error code %q not found in openapi.yaml examples/schemas", code)
	}
}

func TestHTTPResponsesMatchOpenAPI(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)

	t.Setenv("POSTGRES_DSN", suite.DSN(t))
	t.Setenv("ADMIN_API_TOKEN", adminToken)

	cfg, err := config.Load(filepath.Join(openapispec.ProjectRoot(), "config", "config.yaml"))
	require.NoError(t, err)
	cfg.Workers.Enabled = false
	cfg.Swagger.Enabled = false
	cfg.OpenAPI.ValidateRequests = true
	cfg.OpenAPI.ValidateResponses = true

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler, err := app.NewHTTPHandler(context.Background(), cfg, pool, logger)
	require.NoError(t, err)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	t.Run("health_live", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/health/live")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("health_ready", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/health/ready")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("list_venues", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/api/v1/venues?city_id=1")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get_venue_by_slug", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/api/v1/venues/"+venueSlug)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get_menu", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/api/v1/venues/"+venueSlug+"/menu")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("cart_get_empty", func(t *testing.T) {
		resp := customerGET(t, srv.URL+"/api/v1/cart")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("cart_add_and_get", func(t *testing.T) {
		addBody := map[string]any{
			"venue_id":     "11111111-1111-4111-8111-111111111101",
			"menu_item_id": menuItemID,
			"quantity":     2,
			"option_ids":   []string{sizeMedium},
		}
		resp := customerPOST(t, srv.URL+"/api/v1/cart/items", addBody)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		resp2 := customerGET(t, srv.URL+"/api/v1/cart")
		defer resp2.Body.Close()
		require.Equal(t, http.StatusOK, resp2.StatusCode)
	})

	t.Run("partner_unauthorized", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/partner/v1/orders")
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assertErrorCode(t, resp, "unauthorized")
	})

	t.Run("partner_list_orders", func(t *testing.T) {
		resp := partnerGET(t, srv.URL+"/partner/v1/orders")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("admin_unauthorized", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/admin/v1/orders")
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("admin_list_orders", func(t *testing.T) {
		resp := adminGET(t, srv.URL+"/admin/v1/orders")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid_customer_id", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/cart", nil)
		require.NoError(t, err)
		req.Header.Set("X-Customer-Id", "not-a-uuid")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assertErrorCode(t, resp, "invalid_customer_id")
	})

	t.Run("venue_not_found", func(t *testing.T) {
		resp := mustGET(t, srv.URL+"/api/v1/venues/00000000-0000-4000-8000-000000000099")
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		assertErrorCode(t, resp, "venue_not_found")
	})
}

func has4xxResponse(responses *openapi3.Responses) bool {
	if responses == nil {
		return false
	}
	for code := range responses.Map() {
		if strings.HasPrefix(code, "4") {
			return true
		}
	}
	return false
}

func mustGET(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url) //nolint:gosec,noctx // httptest local server
	require.NoError(t, err)
	return resp
}

func customerGET(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("X-Customer-Id", customerID)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func customerPOST(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	require.NoError(t, err)
	req.Header.Set("X-Customer-Id", customerID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func partnerGET(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+partnerAPIKey)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func adminGET(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func assertErrorCode(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(raw, &payload))
	require.Equal(t, want, payload.Error.Code)
}
