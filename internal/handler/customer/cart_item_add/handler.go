package cart_item_add

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	cartadd "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_add"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler {
	return &Handler{uc: uc, validator: v}
}

type addRequest struct {
	VenueID         uuid.UUID       `json:"venue_id" validate:"required"`
	MenuItemID      uuid.UUID       `json:"menu_item_id" validate:"required"`
	Quantity        int             `json:"quantity" validate:"required,min=1,max=99"`
	SelectedOptions json.RawMessage `json:"selected_options"`
	OptionIDs       []uuid.UUID     `json:"option_ids"`
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid, ok := middleware.CustomerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrInvalidCustomerID), nil)
		return
	}
	var req addRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	out, err := h.uc.Run(ctx, cartadd.In{
		CustomerID: cid, VenueID: req.VenueID, MenuItemID: req.MenuItemID,
		Quantity: req.Quantity, SelectedOptions: req.SelectedOptions, OptionIDs: req.OptionIDs,
	})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, map[string]any{"cart": mapper.CartFrom(out.Result)})
}
