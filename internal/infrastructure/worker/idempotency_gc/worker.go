package idempotency_gc

import (
	"context"
	"time"

	idempotencyrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/idempotency"
)

type Worker struct {
	idempotency idempotencyrepo.Repository
}

func New(idempotency idempotencyrepo.Repository) *Worker {
	return &Worker{idempotency: idempotency}
}

func (w *Worker) Name() string { return "idempotency_gc" }

func (w *Worker) Interval() time.Duration { return time.Hour }

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	return w.idempotency.DeleteExpired(ctx)
}
