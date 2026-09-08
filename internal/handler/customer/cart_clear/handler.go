package cart_clear

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	cartclear "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/clear"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler {
	return &Handler{uc: uc, validator: v}
}

func (h *Handler) Clear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid, ok := middleware.CustomerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrInvalidCustomerID), nil)
		return
	}
	if _, err := h.uc.Run(ctx, cartclear.In{CustomerID: cid}); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}
