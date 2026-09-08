package stoplist_restore

import (
	"context"
	"time"

	stoplistrestore "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/stoplist_restore"
)

type Worker struct {
	uc *stoplistrestore.UseCase
}

func New(uc *stoplistrestore.UseCase) *Worker {
	return &Worker{uc: uc}
}

func (w *Worker) Name() string { return "stoplist_restore" }

func (w *Worker) Interval() time.Duration { return time.Minute }

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	out, err := w.uc.Run(ctx, stoplistrestore.In{})
	if err != nil {
		return 0, err
	}
	return out.Restored, nil
}
