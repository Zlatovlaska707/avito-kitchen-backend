package shared

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/money"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type OptionSnapshot struct {
	OptionID        uuid.UUID
	ExternalID      string
	Name            string
	PriceDeltaMinor int64
}

type storedOption struct {
	OptionID        uuid.UUID `json:"option_id"`
	ExternalID      string    `json:"external_id,omitempty"`
	Name            string    `json:"name"`
	PriceDeltaMinor int64     `json:"price_delta_minor"`
}

func BuildOptionsFromIDs(mi catalogrepo.MenuItem, optionIDs []uuid.UUID) ([]OptionSnapshot, json.RawMessage, error) {
	optByID, groupOf := indexMenuOptions(mi)
	selected := make(map[uuid.UUID]int, len(optionIDs))
	seen := make(map[uuid.UUID]struct{}, len(optionIDs))
	snaps := make([]OptionSnapshot, 0, len(optionIDs))

	for _, id := range optionIDs {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		opt, ok := optByID[id]
		if !ok {
			return nil, nil, usecase.ErrInvalidOptions
		}
		if !opt.IsAvailable {
			return nil, nil, usecase.ErrInvalidOptions
		}
		groupID := groupOf[id]
		selected[groupID]++
		snaps = append(snaps, OptionSnapshot{
			OptionID: id, ExternalID: opt.ExternalID, Name: opt.Name, PriceDeltaMinor: opt.PriceDeltaMinor,
		})
	}

	for _, g := range mi.OptionGroups {
		count := selected[g.ID]
		if count < g.MinChoices || count > g.MaxChoices {
			return nil, nil, usecase.ErrInvalidOptions
		}
	}

	raw, err := marshalOptionSnapshots(snaps)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal options: %w", err)
	}
	return snaps, raw, nil
}

func ParseStoredOptions(raw json.RawMessage) ([]OptionSnapshot, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var stored []storedOption
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, err
	}
	snaps := make([]OptionSnapshot, 0, len(stored))
	for _, s := range stored {
		snaps = append(snaps, OptionSnapshot(s))
	}
	return snaps, nil
}

func ResolveLineOptions(mi catalogrepo.MenuItem, raw json.RawMessage) ([]OptionSnapshot, error) {
	snaps, err := ParseStoredOptions(raw)
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return snaps, nil
	}
	optByID, groupOf := indexMenuOptions(mi)
	selected := make(map[uuid.UUID]int, len(snaps))
	for i, s := range snaps {
		opt, ok := optByID[s.OptionID]
		if !ok {
			return nil, usecase.ErrInvalidOptions
		}
		if !opt.IsAvailable {
			return nil, usecase.ErrInvalidOptions
		}
		groupID := groupOf[s.OptionID]
		selected[groupID]++
		snaps[i] = OptionSnapshot{
			OptionID: s.OptionID, ExternalID: opt.ExternalID, Name: opt.Name, PriceDeltaMinor: opt.PriceDeltaMinor,
		}
	}
	for _, g := range mi.OptionGroups {
		count := selected[g.ID]
		if count < g.MinChoices || count > g.MaxChoices {
			return nil, usecase.ErrInvalidOptions
		}
	}
	return snaps, nil
}

func OptionsToRaw(snaps []OptionSnapshot) (json.RawMessage, error) {
	return marshalOptionSnapshots(snaps)
}

func LineUnitPrice(mi catalogrepo.MenuItem, snaps []OptionSnapshot) money.Minor {
	unit := money.Minor(mi.PriceMinor)
	for _, o := range snaps {
		unit += money.Minor(o.PriceDeltaMinor)
	}
	return unit
}

func marshalOptionSnapshots(snaps []OptionSnapshot) (json.RawMessage, error) {
	stored := make([]storedOption, 0, len(snaps))
	for _, s := range snaps {
		stored = append(stored, storedOption(s))
	}
	return json.Marshal(stored)
}

func indexMenuOptions(mi catalogrepo.MenuItem) (map[uuid.UUID]catalogrepo.Option, map[uuid.UUID]uuid.UUID) {
	optByID := make(map[uuid.UUID]catalogrepo.Option)
	groupOf := make(map[uuid.UUID]uuid.UUID)
	for _, g := range mi.OptionGroups {
		for _, o := range g.Options {
			optByID[o.ID] = o
			groupOf[o.ID] = g.ID
		}
	}
	return optByID, groupOf
}
