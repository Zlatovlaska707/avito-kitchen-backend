package webhook_create

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	whcreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/webhook_create"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }

type createReq struct {
	URL        string   `json:"url" validate:"required"`
	Secret     string   `json:"secret"`
	EventTypes []string `json:"event_types" validate:"required,min=1"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, ok := middleware.PartnerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrUnauthorized), nil)
		return
	}
	var req createReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	out, err := h.uc.Run(ctx, whcreate.In{VenueID: p.VenueID, URL: req.URL, Secret: req.Secret, EventTypes: req.EventTypes})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, map[string]any{"id": out.Webhook.ID, "url": out.Webhook.URL, "event_types": out.Webhook.EventTypes})
}
