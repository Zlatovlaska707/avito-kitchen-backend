package menu_sync

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	menusync "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_sync"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }

type syncRequest struct {
	Categories []catalogrepo.SyncCategory `json:"categories" validate:"required"`
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, ok := middleware.PartnerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrUnauthorized), nil)
		return
	}
	var req syncRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	out, err := h.uc.Run(ctx, menusync.In{VenueID: p.VenueID, Categories: req.Categories})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, out.Result)
}
