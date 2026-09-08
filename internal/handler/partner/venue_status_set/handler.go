package venue_status_set

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	statusset "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/status_set"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }

type statusRequest struct {
	AcceptsOrders *bool   `json:"accepts_orders"`
	PauseReason   *string `json:"pause_reason"`
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, ok := middleware.PartnerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrUnauthorized), nil)
		return
	}
	var req statusRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	if _, err := h.uc.Run(ctx, statusset.In{VenueID: p.VenueID, AcceptsOrders: req.AcceptsOrders, PauseReason: req.PauseReason}); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"accepts_orders": req.AcceptsOrders})
}
