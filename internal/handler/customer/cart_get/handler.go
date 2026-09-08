package cart_get

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	cartget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/get"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler {
	return &Handler{uc: uc, validator: v}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid, ok := middleware.CustomerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrInvalidCustomerID), nil)
		return
	}
	out, err := h.uc.Run(ctx, cartget.In{CustomerID: cid})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"cart": mapper.CartFrom(out.Result)})
}
