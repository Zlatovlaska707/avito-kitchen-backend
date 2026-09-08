package order_list

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	orderlist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_venue"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, ok := middleware.PartnerFromContext(ctx)
	if !ok {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrUnauthorized), nil)
		return
	}
	limit, _ := httpx.ParseLimit(r, 20, 100)
	cursorTS, cursorID, _ := httpx.ParseCursor(r)
	var status *string
	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}
	vid := p.VenueID
	out, err := h.uc.Run(ctx, orderlist.In{VenueID: &vid, Status: status, Limit: limit, CursorTS: cursorTS, CursorID: cursorID})
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
