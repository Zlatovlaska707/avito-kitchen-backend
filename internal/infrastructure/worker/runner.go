package worker

import (
	"context"
	"log/slog"
	"time"
)

type Worker interface {
	Name() string
	Interval() time.Duration
	RunOnce(ctx context.Context) (processed int, err error)
}

type Runner struct {
	workers []Worker
	logger  *slog.Logger
	stop    chan struct{}
	done    chan struct{}
}

func NewRunner(logger *slog.Logger, workers ...Worker) *Runner {
	return &Runner{workers: workers, logger: logger, stop: make(chan struct{}), done: make(chan struct{})}
}

func (r *Runner) Start(ctx context.Context) {
	go func() {
		defer close(r.done)
		for _, w := range r.workers {
			w := w
			go r.loop(ctx, w)
		}
		<-r.stop
	}()
}

func (r *Runner) loop(ctx context.Context, w Worker) {
	ticker := time.NewTicker(w.Interval())
	defer ticker.Stop()
	r.runSafe(ctx, w)
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stop:
			return
		case <-ticker.C:
			r.runSafe(ctx, w)
		}
	}
}

func (r *Runner) runSafe(ctx context.Context, w Worker) {
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Error("worker panic", "worker", w.Name(), "panic", rec)
		}
	}()
	start := time.Now()
	n, err := w.RunOnce(ctx)
	if err != nil {
		r.logger.Error("worker error", "worker", w.Name(), "err", err)
		return
	}
	if n > 0 {
		r.logger.Info("worker processed", "worker", w.Name(), "processed", n, "duration_ms", time.Since(start).Milliseconds())
	}
}

func (r *Runner) Stop() {
	close(r.stop)
	<-r.done
}
