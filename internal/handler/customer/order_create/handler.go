package order_create

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	idempotencyrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/idempotency"
	ordercreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/create"
)

type Handler struct {
	uc          UseCase
	validator   *validator.Validate
	idempotency idempotencyrepo.Repository
	logger      *slog.Logger
}

func New(uc UseCase, v *validator.Validate, idempotency idempotencyrepo.Repository, logger *slog.Logger) *Handler {
	return &Handler{uc: uc, validator: v, idempotency: idempotency, logger: logger}
}

type createRequest struct {
	FulfillmentType    string          `json:"fulfillment_type" validate:"required,oneof=delivery pickup"`
	DeliveryAddress    json.RawMessage `json:"address"`
	ContactPhone       string          `json:"contact_phone" validate:"required"`
	PaymentMethod      string          `json:"payment_method" validate:"required,oneof=cash card_online_mock"`
	Comment            *string         `json:"comment"`
	ExpectedTotalMinor int64           `json:"expected_total_minor" validate:"required"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid, ok := middleware.CustomerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrInvalidCustomerID), nil)
		return
	}
	idemKey := r.Header.Get("Idempotency-Key")
	if len(idemKey) < 8 || len(idemKey) > 200 {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), nil)
		return
	}
	var req createRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	hash := requestHash(cid, req)
	out, err := h.uc.Run(ctx, ordercreate.In{
		CustomerID: cid, FulfillmentType: req.FulfillmentType, DeliveryAddress: req.DeliveryAddress,
		ContactPhone: req.ContactPhone, PaymentMethod: req.PaymentMethod, Comment: req.Comment,
		ExpectedTotalMinor: &req.ExpectedTotalMinor, IdempotencyKey: idemKey, RequestHash: hash,
	})
	if err != nil {
		apiErr := apierr.MapDomainError(err)
		if apiErr.Status >= http.StatusInternalServerError && h.logger != nil {
			h.logger.Error("order create failed", "err", err, "request_id", httpx.RequestIDFromContext(ctx))
		}
		httpx.RespondError(w, ctx, apiErr, err)
		return
	}
	if out.Replayed {
		w.Header().Set("Idempotency-Replayed", "true")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(out.ReplayCode)
		_, _ = w.Write(out.ReplayBody)
		return
	}
	order := mapper.OrderFrom(out.Order, nil, nil)
	body, _ := json.Marshal(map[string]any{"order": order})
	_ = h.idempotency.SaveResponse(ctx, idempotencyrepo.SaveResponseIn{
		Scope: "orders:create", Key: idemKey, ResponseStatus: http.StatusCreated,
		ResponseBody: body, ResourceID: &out.Order.ID,
	})
	w.Header().Set("Location", "/api/v1/orders/"+out.Order.ID.String())
	httpx.RespondJSON(w, http.StatusCreated, map[string]any{"order": order})
}

func requestHash(customerID interface{}, req createRequest) string {
	payload, _ := json.Marshal(map[string]any{
		"customer_id": customerID, "body": req,
	})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
