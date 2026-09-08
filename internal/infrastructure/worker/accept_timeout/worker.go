package accept_timeout

import (
	"context"
	"time"

	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	orderexpire "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/expire"
)

type Worker struct {
	orders  orderrepo.Repository
	expire  *orderexpire.UseCase
	timeout time.Duration
}

func New(orders orderrepo.Repository, expire *orderexpire.UseCase, timeout time.Duration) *Worker {
	return &Worker{orders: orders, expire: expire, timeout: timeout}
}

func (w *Worker) Name() string { return "accept_timeout" }

func (w *Worker) Interval() time.Duration { return 30 * time.Second }

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	before := time.Now().Add(-w.timeout)
	ids, err := w.orders.FindPlacedBefore(ctx, before, 100)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, id := range ids {
		out, err := w.expire.Run(ctx, orderexpire.In{OrderID: id})
		if err != nil {
			return n, err
		}
		if out.Processed {
			n++
		}
	}
	return n, nil
}
