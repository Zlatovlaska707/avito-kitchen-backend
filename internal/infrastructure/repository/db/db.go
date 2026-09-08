package db

import (
	"context"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type RepositoryBase struct {
	Pool   *pgxpool.Pool
	Getter *trmpgx.CtxGetter
}

func (b RepositoryBase) Q(ctx context.Context) Querier {
	if b.Getter == nil {
		return b.Pool
	}
	return b.Getter.DefaultTrOrDB(ctx, b.Pool)
}
