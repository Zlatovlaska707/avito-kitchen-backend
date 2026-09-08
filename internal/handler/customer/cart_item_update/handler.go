package cart_item_update

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
	cartupdate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_update"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler {
	return &Handler{uc: uc, validator: v}
}

type updateRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1,max=99"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid, ok := middleware.CustomerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrInvalidCustomerID), nil)
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	var req updateRequest
	if decodeErr := httpx.DecodeJSON(r, &req); decodeErr != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), decodeErr)
		return
	}
	if validateErr := h.validator.Struct(req); validateErr != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), validateErr)
		return
	}
	out, err := h.uc.Run(ctx, cartupdate.In{CustomerID: cid, ItemID: itemID, Quantity: req.Quantity})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"cart": mapper.CartFrom(out.Result)})
}
