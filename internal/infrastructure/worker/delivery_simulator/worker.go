package delivery_simulator

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
	deliveryadvance "github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/advance"
	deliveryassign "github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/assign"
)

var couriers = []struct{ name, phone string }{
	{"Ivan Courier", "+79001112233"},
	{"Maria Delivery", "+79004445566"},
	{"Alex Rider", "+79007778899"},
}

type Worker struct {
	deliveries   deliveryrepo.Repository
	assign       *deliveryassign.UseCase
	advance      *deliveryadvance.UseCase
	assignDelay  time.Duration
	transitDelay time.Duration
	enabled      bool
}

func New(deliveries deliveryrepo.Repository, assign *deliveryassign.UseCase, advance *deliveryadvance.UseCase, assignDelay, transitDelay time.Duration, enabled bool) *Worker {
	return &Worker{
		deliveries: deliveries, assign: assign, advance: advance,
		assignDelay: assignDelay, transitDelay: transitDelay, enabled: enabled,
	}
}

func (w *Worker) Name() string { return "delivery_simulator" }

func (w *Worker) Interval() time.Duration { return 15 * time.Second }

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	if !w.enabled {
		return 0, nil
	}
	n := 0
	assignBefore := time.Now().Add(-w.assignDelay)
	pending, err := w.deliveries.FindPendingForAssign(ctx, assignBefore, 50)
	if err != nil {
		return 0, err
	}
	for _, d := range pending {
		c := couriers[rand.Intn(len(couriers))] //nolint:gosec
		assignOut, assignErr := w.assign.Run(ctx, deliveryassign.In{
			OrderID: d.OrderID, CourierName: c.name, CourierPhone: c.phone,
		})
		if assignErr != nil {
			return n, assignErr
		}
		if assignOut.Assigned {
			n++
		}
	}
	transitBefore := time.Now().Add(-w.transitDelay)
	assigned, err := w.deliveries.FindAssignedForTransit(ctx, transitBefore, 50)
	if err != nil {
		return n, err
	}
	for _, d := range assigned {
		out, err := w.advance.Run(ctx, deliveryadvance.In{OrderID: d.OrderID})
		if err != nil {
			return n, fmt.Errorf("advance delivery: %w", err)
		}
		if out.Completed {
			n++
		}
	}
	return n, nil
}
