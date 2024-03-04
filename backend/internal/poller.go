package internal

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Poller struct {
	producer *Producer
	postgres *pgxpool.Pool
	period   time.Duration
}

func NewPoller(producer *Producer, postgres *pgxpool.Pool, period time.Duration) *Poller {
	return &Poller{
		producer: producer,
		postgres: postgres,
		period:   period,
	}
}

func (p *Poller) Run(ctx context.Context) {
	p.poll(ctx)

	ticker := time.NewTicker(p.period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.poll(ctx)

		case <-ctx.Done():
			return
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	slog.Info("Polling for tasks")
	rows, err := p.postgres.Query(ctx, "SELECT id, prompt FROM caption_tasks WHERE status = 'created' AND created_at < NOW() - INTERVAL '1 minute' FOR UPDATE SKIP LOCKED")
	if err != nil {
		slog.Error("Failed to query tasks", slog.String("error", err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task CaptionTaskEvent
		err = rows.Scan(&task.Id, &task.Prompt)

		if err != nil {
			slog.Error("Failed to scan task", slog.String("error", err.Error()))
			continue
		}

		err = p.producer.SendTaskEvent(&task)
		if err != nil {
			slog.Error("Failed to send task", slog.String("error", err.Error()))
			continue
		}
	}
	slog.Info("Finished polling for tasks")
}
