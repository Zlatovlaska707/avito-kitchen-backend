package shared_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

func sampleMenuItem() catalogrepo.MenuItem {
	groupID := uuid.New()
	optID := uuid.New()
	return catalogrepo.MenuItem{
		ID:          uuid.New(),
		PriceMinor:  49000,
		OptionGroups: []catalogrepo.OptionGroup{{
			ID: groupID, MinChoices: 1, MaxChoices: 1,
			Options: []catalogrepo.Option{{
				ID: optID, ExternalID: "size-m", Name: "Medium", PriceDeltaMinor: 10000, IsAvailable: true,
			}},
		}},
	}
}

func TestBuildOptionsFromIDs_HappyPath(t *testing.T) {
	mi := sampleMenuItem()
	optID := mi.OptionGroups[0].Options[0].ID
	snaps, raw, err := shared.BuildOptionsFromIDs(mi, []uuid.UUID{optID})
	require.NoError(t, err)
	require.Len(t, snaps, 1)
	assert.Equal(t, int64(10000), snaps[0].PriceDeltaMinor)
	assert.NotEmpty(t, raw)
}

func TestBuildOptionsFromIDs_InvalidOption(t *testing.T) {
	mi := sampleMenuItem()
	_, _, err := shared.BuildOptionsFromIDs(mi, []uuid.UUID{uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrInvalidOptions)
}

func TestBuildOptionsFromIDs_UnavailableOption(t *testing.T) {
	mi := sampleMenuItem()
	mi.OptionGroups[0].Options[0].IsAvailable = false
	optID := mi.OptionGroups[0].Options[0].ID
	_, _, err := shared.BuildOptionsFromIDs(mi, []uuid.UUID{optID})
	assert.ErrorIs(t, err, usecase.ErrInvalidOptions)
}

func TestBuildOptionsFromIDs_MinChoices(t *testing.T) {
	mi := sampleMenuItem()
	_, _, err := shared.BuildOptionsFromIDs(mi, nil)
	assert.ErrorIs(t, err, usecase.ErrInvalidOptions)
}

func TestParseStoredOptions(t *testing.T) {
	snaps, err := shared.ParseStoredOptions(nil)
	require.NoError(t, err)
	assert.Nil(t, snaps)

	raw := json.RawMessage(`[{"option_id":"00000000-0000-0000-0000-000000000001","name":"S","price_delta_minor":0}]`)
	snaps, err = shared.ParseStoredOptions(raw)
	require.NoError(t, err)
	require.Len(t, snaps, 1)
}

func TestResolveLineOptions(t *testing.T) {
	mi := sampleMenuItem()
	optID := mi.OptionGroups[0].Options[0].ID
	raw, err := json.Marshal([]map[string]any{
		{"option_id": optID, "name": "old", "price_delta_minor": 0},
	})
	require.NoError(t, err)
	snaps, err := shared.ResolveLineOptions(mi, raw)
	require.NoError(t, err)
	require.Len(t, snaps, 1)
	assert.Equal(t, "Medium", snaps[0].Name)
}

func TestLineUnitPrice(t *testing.T) {
	mi := sampleMenuItem()
	price := shared.LineUnitPrice(mi, []shared.OptionSnapshot{{PriceDeltaMinor: 5000}})
	assert.Equal(t, int64(54000), int64(price))
}

func TestOptionsToRaw(t *testing.T) {
	raw, err := shared.OptionsToRaw([]shared.OptionSnapshot{{Name: "x", PriceDeltaMinor: 1}})
	require.NoError(t, err)
	assert.Contains(t, string(raw), "x")
}
