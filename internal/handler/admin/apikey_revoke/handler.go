package apikey_revoke

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	apikeyrevoke "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/apikey_revoke"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vid, err := uuid.Parse(chi.URLParam(r, "venueId"))
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	kid, err := uuid.Parse(chi.URLParam(r, "keyId"))
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	if _, err := h.uc.Run(ctx, apikeyrevoke.In{VenueID: vid, KeyID: kid}); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}
