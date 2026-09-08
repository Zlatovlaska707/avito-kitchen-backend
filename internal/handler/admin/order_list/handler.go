package order_list

import (
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	orderlist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_admin"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, _ := httpx.ParseLimit(r, 20, 100)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var venueID *uuid.UUID
	if v := r.URL.Query().Get("venue_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			venueID = &id
		}
	}
	out, err := h.uc.Run(ctx, orderlist.In{VenueID: venueID, Limit: limit, Offset: offset})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	items := make([]mapper.Order, 0, len(out.Orders))
	for _, o := range out.Orders {
		oo := o
		items = append(items, mapper.OrderFrom(&oo, nil, nil))
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"items": items, "page_info": mapper.PageFrom(limit, out.HasNext, out.LastTS, out.LastID)})
}
