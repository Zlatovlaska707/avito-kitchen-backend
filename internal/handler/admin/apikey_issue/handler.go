package apikey_issue

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	apikeyissue "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/apikey_issue"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }

type issueReq struct {
	Scopes []string `json:"scopes"`
}

func (h *Handler) Issue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	var req issueReq
	_ = httpx.DecodeJSON(r, &req)
	out, err := h.uc.Run(ctx, apikeyissue.In{VenueID: vid, Scopes: req.Scopes})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, map[string]any{"key_id": out.KeyID, "api_key": out.APIKey, "prefix": out.Prefix})
}
