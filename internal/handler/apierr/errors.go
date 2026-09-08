package apierr

import (
	"errors"
	"net/http"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
)

func MapDomainError(err error) httpx.APIError {
	switch {
	case errors.Is(err, usecase.ErrUnauthorized):
		return httpx.APIError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: "Unauthorized"}
	case errors.Is(err, usecase.ErrForbidden):
		return httpx.APIError{Status: http.StatusForbidden, Code: "forbidden", Message: "Forbidden"}
	case errors.Is(err, usecase.ErrValidation):
		return withDetails(http.StatusBadRequest, "validation_failed", "Validation failed", err)
	case errors.Is(err, usecase.ErrInvalidCustomerID):
		return httpx.APIError{Status: http.StatusBadRequest, Code: "invalid_customer_id", Message: "Invalid customer id"}
	case errors.Is(err, usecase.ErrVenueNotFound):
		return httpx.APIError{Status: http.StatusNotFound, Code: "venue_not_found", Message: "Venue not found"}
	case errors.Is(err, usecase.ErrVenueSlugTaken):
		return httpx.APIError{Status: http.StatusConflict, Code: "venue_slug_taken", Message: "Venue slug already taken"}
	case errors.Is(err, usecase.ErrVenueClosed):
		return withDetails(http.StatusConflict, "venue_closed", "Venue is closed", err)
	case errors.Is(err, usecase.ErrMenuItemNotFound):
		return httpx.APIError{Status: http.StatusNotFound, Code: "menu_item_not_found", Message: "Menu item not found"}
	case errors.Is(err, usecase.ErrMenuItemUnavailable):
		return withDetails(http.StatusConflict, "item_unavailable", "Some items are no longer available", err)
	case errors.Is(err, usecase.ErrInvalidOptions):
		return httpx.APIError{Status: http.StatusBadRequest, Code: "invalid_options", Message: "Invalid options"}
	case errors.Is(err, usecase.ErrDuplicateExternalID):
		return withDetails(http.StatusBadRequest, "duplicate_external_id", "Duplicate external id in payload", err)
	case errors.Is(err, usecase.ErrMenuSyncInProgress):
		return httpx.APIError{Status: http.StatusConflict, Code: "menu_sync_in_progress", Message: "Menu sync already in progress"}
	case errors.Is(err, usecase.ErrCartNotFound):
		return httpx.APIError{Status: http.StatusNotFound, Code: "cart_not_found", Message: "Cart not found"}
	case errors.Is(err, usecase.ErrCartEmpty):
		return httpx.APIError{Status: http.StatusConflict, Code: "cart_empty", Message: "Cart is empty"}
	case errors.Is(err, usecase.ErrCartVenueMismatch):
		return withDetails(http.StatusConflict, "cart_venue_mismatch", "Cart belongs to another venue", err)
	case errors.Is(err, usecase.ErrCartItemLimit):
		return httpx.APIError{Status: http.StatusUnprocessableEntity, Code: "cart_item_limit_exceeded", Message: "Cart item limit exceeded"}
	case errors.Is(err, usecase.ErrQuantityLimit):
		return httpx.APIError{Status: http.StatusUnprocessableEntity, Code: "quantity_limit_exceeded", Message: "Quantity limit exceeded"}
	case errors.Is(err, usecase.ErrBelowMinOrder):
		return withDetails(http.StatusConflict, "below_min_order", "Order total below minimum", err)
	case errors.Is(err, usecase.ErrOutsideDeliveryZone):
		return withDetails(http.StatusConflict, "outside_delivery_zone", "Address outside delivery zone", err)
	case errors.Is(err, usecase.ErrPriceChanged):
		return withDetails(http.StatusConflict, "price_changed", "Price has changed", err)
	case errors.Is(err, usecase.ErrOrderNotFound):
		return httpx.APIError{Status: http.StatusNotFound, Code: "order_not_found", Message: "Order not found"}
	case errors.Is(err, usecase.ErrInvalidTransition):
		return withDetails(http.StatusConflict, "invalid_transition", "Invalid status transition", err)
	case errors.Is(err, usecase.ErrNotCancellable):
		return httpx.APIError{Status: http.StatusConflict, Code: "not_cancellable", Message: "Order cannot be cancelled"}
	case errors.Is(err, usecase.ErrIdempotencyKeyReuse):
		return httpx.APIError{Status: http.StatusConflict, Code: "idempotency_key_reuse", Message: "Idempotency key reuse with different body"}
	case errors.Is(err, usecase.ErrRequestInProgress):
		return httpx.APIError{Status: http.StatusConflict, Code: "request_in_progress", Message: "Request with the same idempotency key is in progress"}
	case errors.Is(err, usecase.ErrWebhookURLForbidden):
		return httpx.APIError{Status: http.StatusBadRequest, Code: "webhook_url_forbidden", Message: "Webhook URL is forbidden"}
	case errors.Is(err, usecase.ErrWebhookAlreadyExists):
		return httpx.APIError{Status: http.StatusConflict, Code: "webhook_already_exists", Message: "Webhook already exists"}
	default:
		return httpx.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "Internal server error"}
	}
}

func withDetails(status int, code, message string, err error) httpx.APIError {
	apiErr := httpx.APIError{Status: status, Code: code, Message: message}
	if details := extractDetails(err); len(details) > 0 {
		apiErr.Details = details
	}
	return apiErr
}

func extractDetails(err error) map[string]any {
	details := map[string]any{}
	if err == nil {
		return details
	}
	// Detail structs may be embedded via %v in wrapped errors; avoid errors.As on non-error types.
	type detailError interface {
		Detail() map[string]any
	}
	var de detailError
	if errors.As(err, &de) {
		return de.Detail()
	}
	return details
}
