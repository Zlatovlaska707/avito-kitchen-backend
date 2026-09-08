package uctest

import (
	"context"
	"time"

	"github.com/google/uuid"

	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/webhook"
)

type FixedNower struct{ T time.Time }

func (n FixedNower) Now() time.Time { return n.T }

type OrderRepo struct {
	Order            *orderrepo.Order
	GetByIDErr       error
	GetByCustomerErr error
	ListOrders       []orderrepo.Order
	ListErr          error
	Items            []orderrepo.Item
	ItemsErr         error
	History          []orderrepo.StatusEvent
	HistoryErr       error
	TransitionOut    *orderrepo.Order
	TransitionErr    error
	PlacedBefore     []uuid.UUID
}

func (m *OrderRepo) Create(context.Context, orderrepo.CreateIn) (*orderrepo.Order, error) {
	return nil, nil
}

func (m *OrderRepo) GetByID(context.Context, uuid.UUID) (*orderrepo.Order, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	return m.Order, nil
}

func (m *OrderRepo) GetByIDForCustomer(_ context.Context, _, _ uuid.UUID) (*orderrepo.Order, error) {
	if m.GetByCustomerErr != nil {
		return nil, m.GetByCustomerErr
	}
	return m.Order, nil
}

func (m *OrderRepo) GetByIDForVenue(_ context.Context, _, _ uuid.UUID) (*orderrepo.Order, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	return m.Order, nil
}

func (m *OrderRepo) List(context.Context, orderrepo.ListFilter) ([]orderrepo.Order, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	return m.ListOrders, nil
}

func (m *OrderRepo) GetItems(context.Context, uuid.UUID) ([]orderrepo.Item, error) {
	if m.ItemsErr != nil {
		return nil, m.ItemsErr
	}
	return m.Items, nil
}

func (m *OrderRepo) GetStatusHistory(context.Context, uuid.UUID) ([]orderrepo.StatusEvent, error) {
	if m.HistoryErr != nil {
		return nil, m.HistoryErr
	}
	return m.History, nil
}

func (m *OrderRepo) ApplyTransition(context.Context, orderrepo.TransitionIn) (*orderrepo.Order, error) {
	if m.TransitionErr != nil {
		return nil, m.TransitionErr
	}
	if m.TransitionOut != nil {
		return m.TransitionOut, nil
	}
	return m.Order, nil
}

func (m *OrderRepo) FindPlacedBefore(context.Context, time.Time, int) ([]uuid.UUID, error) {
	return m.PlacedBefore, nil
}

type VenueRepo struct {
	Venue        *venuerepo.Venue
	Venues       []venuerepo.Venue
	GetByIDErr   error
	GetBySlugErr error
	ListErr      error
	Hours        []venuerepo.WorkingHour
	SlugTaken    bool
	Created      *venuerepo.Venue
	Updated      *venuerepo.Venue
	CreateErr    error
	UpdateErr    error
	StatusErr    error
}

func (m *VenueRepo) Create(context.Context, venuerepo.CreateIn) (*venuerepo.Venue, error) {
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	if m.Created != nil {
		return m.Created, nil
	}
	return m.Venue, nil
}

func (m *VenueRepo) Update(context.Context, venuerepo.UpdateIn) (*venuerepo.Venue, error) {
	if m.UpdateErr != nil {
		return nil, m.UpdateErr
	}
	if m.Updated != nil {
		return m.Updated, nil
	}
	return m.Venue, nil
}

func (m *VenueRepo) GetByID(context.Context, uuid.UUID) (*venuerepo.Venue, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	return m.Venue, nil
}

func (m *VenueRepo) GetBySlug(context.Context, string) (*venuerepo.Venue, error) {
	if m.GetBySlugErr != nil {
		return nil, m.GetBySlugErr
	}
	return m.Venue, nil
}

func (m *VenueRepo) List(context.Context, venuerepo.ListFilter) ([]venuerepo.Venue, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	return m.Venues, nil
}

func (m *VenueRepo) UpdateStatus(context.Context, venuerepo.StatusUpdateIn) error {
	return m.StatusErr
}

func (m *VenueRepo) GetWorkingHours(context.Context, uuid.UUID) ([]venuerepo.WorkingHour, error) {
	return m.Hours, nil
}

func (m *VenueRepo) ReplaceWorkingHours(context.Context, uuid.UUID, []venuerepo.WorkingHour) error {
	return nil
}

func (m *VenueRepo) IsSlugTaken(context.Context, string, *uuid.UUID) (bool, error) {
	return m.SlugTaken, nil
}

type DeliveryRepo struct {
	Delivery  *deliveryrepo.Delivery
	UpdateErr error
	UpdateOut *deliveryrepo.Delivery
}

func (m *DeliveryRepo) Create(context.Context, deliveryrepo.CreateIn) (*deliveryrepo.Delivery, error) {
	return nil, nil
}

func (m *DeliveryRepo) UpdateStatus(context.Context, deliveryrepo.UpdateStatusIn) (*deliveryrepo.Delivery, error) {
	if m.UpdateErr != nil {
		return nil, m.UpdateErr
	}
	if m.UpdateOut != nil {
		return m.UpdateOut, nil
	}
	return m.Delivery, nil
}

func (m *DeliveryRepo) GetByOrderID(context.Context, uuid.UUID) (*deliveryrepo.Delivery, error) {
	return m.Delivery, nil
}

func (m *DeliveryRepo) FindPendingForAssign(context.Context, time.Time, int) ([]deliveryrepo.Delivery, error) {
	return nil, nil
}

func (m *DeliveryRepo) FindAssignedForTransit(context.Context, time.Time, int) ([]deliveryrepo.Delivery, error) {
	return nil, nil
}

type WebhookRepo struct {
	Webhook   *webhookrepo.Webhook
	CreateErr error
	DeleteErr error
}

func (m *WebhookRepo) Create(context.Context, webhookrepo.CreateIn) (*webhookrepo.Webhook, error) {
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	return m.Webhook, nil
}

func (m *WebhookRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	return m.DeleteErr
}

func (m *WebhookRepo) GetByID(context.Context, uuid.UUID, uuid.UUID) (*webhookrepo.Webhook, error) {
	return m.Webhook, nil
}

func (m *WebhookRepo) ListByVenue(context.Context, uuid.UUID) ([]webhookrepo.Webhook, error) {
	return nil, nil
}

func (m *WebhookRepo) FindActiveByVenueAndEvent(context.Context, uuid.UUID, string) ([]webhookrepo.Webhook, error) {
	return nil, nil
}
