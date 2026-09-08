package order_reject

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	ucreject "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/reject"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }

type rejectReq struct {
	Reason     string `json:"reason" validate:"required"`
	ReasonCode string `json:"reason_code"`
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, ok := middleware.PartnerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrUnauthorized), nil)
		return
	}
	oid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	var req rejectReq
	if decodeErr := httpx.DecodeJSON(r, &req); decodeErr != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), decodeErr)
		return
	}
	out, err := h.uc.Run(ctx, ucreject.In{OrderID: oid, VenueID: p.VenueID, Reason: req.Reason})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"order": mapper.OrderFrom(out.Order, nil, nil)})
}
