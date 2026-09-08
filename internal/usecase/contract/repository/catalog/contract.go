package catalog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetMenu(ctx context.Context, venueID uuid.UUID) (*Menu, string, error)
	SyncMenu(ctx context.Context, in SyncIn) (*SyncResult, error)
	SetItemAvailability(ctx context.Context, in AvailabilityIn) error
	GetItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]MenuItem, error)
	TryMenuSyncLock(ctx context.Context, venueID uuid.UUID) (bool, error)
	RestoreExpiredStoplist(ctx context.Context) (int, error)
}

type Menu struct {
	VenueID    uuid.UUID
	Categories []Category
	UpdatedAt  time.Time
}

type Category struct {
	ID         uuid.UUID
	ExternalID string
	Name       string
	SortOrder  int
	IsActive   bool
	Items      []MenuItem
}

type MenuItem struct {
	ID               uuid.UUID
	VenueID          uuid.UUID
	CategoryID       uuid.UUID
	ExternalID       string
	Name             string
	Description      *string
	PriceMinor       int64
	WeightGrams      *int
	ImageURL         *string
	SortOrder        int
	IsActive         bool
	IsAvailable      bool
	UnavailableUntil *time.Time
	Revision         int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	OptionGroups     []OptionGroup
}

type OptionGroup struct {
	ID         uuid.UUID
	ExternalID string
	Name       string
	MinChoices int
	MaxChoices int
	SortOrder  int
	Options    []Option
}

type Option struct {
	ID              uuid.UUID
	ExternalID      string
	Name            string
	PriceDeltaMinor int64
	IsAvailable     bool
	SortOrder       int
}

type SyncIn struct {
	VenueID    uuid.UUID
	Categories []SyncCategory
}

type SyncCategory struct {
	ExternalID string     `json:"external_id"`
	Name       string     `json:"name"`
	SortOrder  int        `json:"sort_order"`
	Items      []SyncItem `json:"items"`
}

type SyncItem struct {
	ExternalID  string            `json:"external_id"`
	CategoryExt string            `json:"category_external_id,omitempty"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	PriceMinor  int64             `json:"price_minor"`
	WeightGrams *int              `json:"weight_grams,omitempty"`
	ImageURL    *string           `json:"image_url,omitempty"`
	SortOrder   int               `json:"sort_order"`
	Groups      []SyncOptionGroup `json:"option_groups"`
}

type SyncOptionGroup struct {
	ExternalID string       `json:"external_id"`
	Name       string       `json:"name"`
	MinChoices int          `json:"min_choices"`
	MaxChoices int          `json:"max_choices"`
	SortOrder  int          `json:"sort_order"`
	Options    []SyncOption `json:"options"`
}

type SyncOption struct {
	ExternalID      string `json:"external_id"`
	Name            string `json:"name"`
	PriceDeltaMinor int64  `json:"price_delta_minor"`
	SortOrder       int    `json:"sort_order"`
}

type SyncResult struct {
	Created     int
	Updated     int
	Deactivated int
	Unchanged   int
}

type AvailabilityIn struct {
	VenueID          uuid.UUID
	ExternalID       string
	IsAvailable      bool
	UnavailableUntil *time.Time
}

type OptionsSnapshot json.RawMessage
