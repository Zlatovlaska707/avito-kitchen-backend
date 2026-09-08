package venue_update

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venueupdate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/update"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	var req venueupdate.In
	if decodeErr := httpx.DecodeJSON(r, &req); decodeErr != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), decodeErr)
		return
	}
	req.ID = id
	out, err := h.uc.Run(ctx, req)
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"id": out.Venue.ID, "name": out.Venue.Name})
}
