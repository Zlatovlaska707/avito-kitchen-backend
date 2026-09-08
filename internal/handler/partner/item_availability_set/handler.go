package item_availability_set

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	avail "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/item_availability_set"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }

type availRequest struct {
	IsAvailable      bool       `json:"is_available"`
	UnavailableUntil *time.Time `json:"unavailable_until"`
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, ok := middleware.PartnerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrUnauthorized), nil)
		return
	}
	extID := chi.URLParam(r, "externalId")
	var req availRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	if _, err := h.uc.Run(ctx, avail.In{VenueID: p.VenueID, ExternalID: extID, IsAvailable: req.IsAvailable, UnavailableUntil: req.UnavailableUntil}); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"external_id": extID, "is_available": req.IsAvailable})
}
