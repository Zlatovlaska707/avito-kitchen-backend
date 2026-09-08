package cart_gc

import (
	"context"
	"time"

	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
)

type Worker struct {
	carts cartrepo.Repository
}

func New(carts cartrepo.Repository) *Worker {
	return &Worker{carts: carts}
}

func (w *Worker) Name() string { return "cart_gc" }

func (w *Worker) Interval() time.Duration { return 10 * time.Minute }

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	return w.carts.AbandonExpired(ctx)
}
