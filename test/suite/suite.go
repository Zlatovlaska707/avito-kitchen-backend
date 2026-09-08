package suite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trmmanager "github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/nower"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/catalog"
	idempotencyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/idempotency"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/order"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/venue"
	cartitemadd "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_add"
	catalogmenusync "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_sync"
	catalogcontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	ordercreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/create"
)

var (
	once        sync.Once
	pool        *pgxpool.Pool
	container   testcontainers.Container
	postgresDSN string
	initErr     error
)

const (
	PostgresUser = "kitchen"
	PostgresPass = "kitchen"
	PostgresDB   = "kitchen_test"
)

var (
	CustomerID         = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	VenuePizzaMario    = uuid.MustParse("11111111-1111-4111-8111-111111111101")
	MenuItemMargherita = uuid.MustParse("33333333-3333-4333-8333-333333333301")
	SizeMedium         = uuid.MustParse("66666666-6666-4666-8666-666666666302")
)

// Pool returns a shared pgx pool backed by one postgres:16-alpine container per test run.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	once.Do(func() {
		ctx := context.Background()
		pg, err := postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase(PostgresDB),
			postgres.WithUsername(PostgresUser),
			postgres.WithPassword(PostgresPass),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			initErr = fmt.Errorf("start postgres container: %w", err)
			return
		}
		container = pg
		connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			initErr = fmt.Errorf("connection string: %w", err)
			return
		}
		if err := runMigrations(ctx, connStr); err != nil {
			initErr = err
			return
		}
		postgresDSN = connStr
		pool, initErr = pgxpool.New(ctx, connStr)
	})
	require.NoError(t, initErr)
	return pool
}

// DSN returns the shared testcontainers postgres connection string.
func DSN(t *testing.T) string {
	t.Helper()
	_ = Pool(t)
	require.NotEmpty(t, postgresDSN)
	return postgresDSN
}

// Reset truncates application tables and reapplies seed.sql.
func Reset(t *testing.T, p *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := p.Exec(ctx, truncateSQL)
	require.NoError(t, err)
	require.NoError(t, applySeed(ctx, p))
}

// Fixtures wires repositories and key use cases against the shared pool.
type Fixtures struct {
	Pool        *pgxpool.Pool
	CartAdd     *cartitemadd.UseCase
	OrderCreate *ordercreate.UseCase
	MenuSync    *catalogmenusync.UseCase
	Venues      *venuerepo.Repository
	Catalog     *catalogrepo.Repository
}

func NewFixtures(t *testing.T, p *pgxpool.Pool) *Fixtures {
	t.Helper()
	getter := trmpgx.DefaultCtxGetter
	trManager := trmmanager.Must(trmpgx.NewDefaultFactory(p))
	defTZ := mustLoadTZ(t, "Europe/Moscow")
	now := nower.NewReal()

	carts := cartrepo.New(p, getter)
	catalog := catalogrepo.New(p, getter)
	venues := venuerepo.New(p, getter)
	orders := orderrepo.New(p, getter)
	idempotency := idempotencyrepo.New(p, getter)

	return &Fixtures{
		Pool:        p,
		CartAdd:     cartitemadd.New(carts, catalog, venues, now, defTZ, 24*time.Hour, 100),
		OrderCreate: ordercreate.New(trManager, carts, catalog, venues, orders, idempotency, now, defTZ, 24*time.Hour, 10, 3.0),
		MenuSync:    catalogmenusync.New(catalog, 200, 2000),
		Venues:      venues,
		Catalog:     catalog,
	}
}

func mustLoadTZ(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}

func runMigrations(ctx context.Context, conn string) error {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err = goose.SetDialect("postgres"); err != nil {
		return err
	}
	dir, err := migrationsDir()
	if err != nil {
		return err
	}
	if err := goose.UpContext(ctx, db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return applySeedDB(ctx, db)
}

func applySeedDB(ctx context.Context, db *sql.DB) error {
	content, err := os.ReadFile(filepath.Join(migrationsRoot(), "migrations", "seed", "seed.sql"))
	if err != nil {
		return fmt.Errorf("read seed: %w", err)
	}
	if _, err := db.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("apply seed: %w", err)
	}
	return nil
}

func applySeed(ctx context.Context, p *pgxpool.Pool) error {
	content, err := os.ReadFile(filepath.Join(migrationsRoot(), "migrations", "seed", "seed.sql"))
	if err != nil {
		return fmt.Errorf("read seed: %w", err)
	}
	if _, err := p.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("apply seed: %w", err)
	}
	return nil
}

func migrationsDir() (string, error) {
	dir := filepath.Join(migrationsRoot(), "migrations")
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("migrations dir: %w", err)
	}
	return dir, nil
}

func migrationsRoot() string {
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		return root
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd
		}
		dir = parent
	}
}

const truncateSQL = `
TRUNCATE TABLE
	webhook_deliveries,
	outbox_events,
	deliveries,
	idempotency_keys,
	order_status_history,
	order_items,
	orders,
	cart_items,
	carts,
	customers,
	menu_item_options,
	menu_item_option_groups,
	menu_items,
	menu_categories,
	venue_webhooks,
	venue_api_keys,
	venue_working_hours,
	venues
RESTART IDENTITY CASCADE;
`

// Close terminates the shared container. Safe to call multiple times.
func Close() {
	if pool != nil {
		pool.Close()
		pool = nil
	}
	if container != nil {
		_ = container.Terminate(context.Background())
		container = nil
	}
}

// SeedMenuSyncPayload returns a minimal menu payload for integration tests.
func SeedMenuSyncPayload() []catalogcontract.SyncCategory {
	return []catalogcontract.SyncCategory{{
		ExternalID: "cat-test",
		Name:       "Test Category",
		Items: []catalogcontract.SyncItem{{
			ExternalID: "item-test-1",
			Name:       "Test Item",
			PriceMinor: 10000,
		}},
	}}
}

func init() {
	stdlib.GetDefaultDriver()
}
