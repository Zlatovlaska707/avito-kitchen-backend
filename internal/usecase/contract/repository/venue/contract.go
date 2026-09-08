package venue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, in CreateIn) (*Venue, error)
	Update(ctx context.Context, in UpdateIn) (*Venue, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Venue, error)
	GetBySlug(ctx context.Context, slug string) (*Venue, error)
	List(ctx context.Context, filter ListFilter) ([]Venue, error)
	UpdateStatus(ctx context.Context, in StatusUpdateIn) error
	GetWorkingHours(ctx context.Context, venueID uuid.UUID) ([]WorkingHour, error)
	ReplaceWorkingHours(ctx context.Context, venueID uuid.UUID, hours []WorkingHour) error
	IsSlugTaken(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error)
}

type Venue struct {
	ID               uuid.UUID
	Slug             string
	Name             string
	Description      *string
	Status           string
	AcceptsOrders    bool
	CityID           int
	AddressText      string
	Lat              float64
	Lon              float64
	Timezone         string
	AvgPrepMinutes   int
	DeliveryFeeMinor int64
	MinOrderMinor    int64
	DeliveryRadiusM  int
	Rating           *float64
	PauseReason      *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type WorkingHour struct {
	ID              uuid.UUID
	Weekday         int
	OpensAt         time.Duration
	ClosesAt        time.Duration
	CrossesMidnight bool
}

type CreateIn struct {
	Slug             string
	Name             string
	Description      *string
	CityID           int
	AddressText      string
	Lat              float64
	Lon              float64
	Timezone         string
	AvgPrepMinutes   int
	DeliveryFeeMinor int64
	MinOrderMinor    int64
	DeliveryRadiusM  int
	WorkingHours     []WorkingHour
}

type UpdateIn struct {
	ID               uuid.UUID
	Name             *string
	Description      *string
	AddressText      *string
	Lat              *float64
	Lon              *float64
	Timezone         *string
	AvgPrepMinutes   *int
	DeliveryFeeMinor *int64
	MinOrderMinor    *int64
	DeliveryRadiusM  *int
}

type StatusUpdateIn struct {
	VenueID       uuid.UUID
	Status        *string
	AcceptsOrders *bool
	PauseReason   *string
}

type ListFilter struct {
	CityID   *int
	Status   *string
	Query    *string
	Lat      *float64
	Lon      *float64
	RadiusM  *int
	Sort     string
	Limit    int
	CursorTS *time.Time
	CursorID *uuid.UUID
}
