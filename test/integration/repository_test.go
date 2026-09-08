//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_sync"
	catalogcontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	"github.com/zlatovlaska707/avito-kitchen/test/suite"
)

func TestRepository_VenueGetBySlug(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)
	fx := suite.NewFixtures(t, pool)
	ctx := context.Background()

	venue, err := fx.Venues.GetBySlug(ctx, "pizza-mario")
	require.NoError(t, err)
	require.NotNil(t, venue)
	assert.Equal(t, suite.VenuePizzaMario, venue.ID)
	assert.Equal(t, "active", venue.Status)
	assert.True(t, venue.AcceptsOrders)
}

func TestRepository_CatalogGetMenu(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)
	fx := suite.NewFixtures(t, pool)
	ctx := context.Background()

	menu, etag, err := fx.Catalog.GetMenu(ctx, suite.VenuePizzaMario)
	require.NoError(t, err)
	require.NotNil(t, menu)
	assert.NotEmpty(t, etag)
	assert.NotEmpty(t, menu.Categories)

	var itemCount int
	for _, cat := range menu.Categories {
		itemCount += len(cat.Items)
	}
	assert.GreaterOrEqual(t, itemCount, 12)
}

func TestRepository_MenuSync_Idempotent(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)
	fx := suite.NewFixtures(t, pool)
	ctx := context.Background()

	payload := suite.SeedMenuSyncPayload()
	out1, err := fx.MenuSync.Run(ctx, menu_sync.In{VenueID: suite.VenuePizzaMario, Categories: payload})
	require.NoError(t, err)
	assert.Equal(t, 2, out1.Result.Created)

	out2, err := fx.MenuSync.Run(ctx, menu_sync.In{VenueID: suite.VenuePizzaMario, Categories: payload})
	require.NoError(t, err)
	assert.Equal(t, 0, out2.Result.Created)

	_, etag1, err := fx.Catalog.GetMenu(ctx, suite.VenuePizzaMario)
	require.NoError(t, err)
	_, etag2, err := fx.Catalog.GetMenu(ctx, suite.VenuePizzaMario)
	require.NoError(t, err)
	assert.Equal(t, etag1, etag2)
}

func TestRepository_MenuSync_SimulatorPayload(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)
	fx := suite.NewFixtures(t, pool)
	ctx := context.Background()

	content, err := os.ReadFile(filepath.Join("..", "..", "venue-simulator", "menu.json"))
	require.NoError(t, err)
	var req struct {
		Categories []catalogcontract.SyncCategory `json:"categories"`
	}
	require.NoError(t, json.Unmarshal(content, &req))

	out, err := fx.MenuSync.Run(ctx, menu_sync.In{VenueID: suite.VenuePizzaMario, Categories: req.Categories})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, out.Result.Created+out.Result.Updated, 1)

	out2, err := fx.MenuSync.Run(ctx, menu_sync.In{VenueID: suite.VenuePizzaMario, Categories: req.Categories})
	require.NoError(t, err)
	assert.Equal(t, 0, out2.Result.Created)
}
