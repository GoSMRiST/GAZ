package worker

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type EmailCleanupWorker struct {
	log *slog.Logger
	db  *pgxpool.Pool

	interval time.Duration
}

func NewEmailCleanupWorker(log *slog.Logger, db *pgxpool.Pool) *EmailCleanupWorker {
	return &EmailCleanupWorker{
		log:      log,
		db:       db,
		interval: 1 * time.Hour,
	}
}

func (w *EmailCleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)

	w.log.Info("email cleanup worker started", "interval", w.interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.cleanup(ctx)

			case <-ctx.Done():
				w.log.Info("email cleanup worker stopped")
				return
			}
		}
	}()
}

func (w *EmailCleanupWorker) cleanup(ctx context.Context) {
	result, err := w.db.Exec(ctx, `
        DELETE FROM email_verifications WHERE expires_at < NOW()
    `)
	if err != nil {
		w.log.Error("email_verifications cleanup failed", "error", err)
	} else {
		w.log.Info("email_verifications cleanup", "deleted_rows", result.RowsAffected())
	}

	result, err = w.db.Exec(ctx, `
        DELETE FROM password_resets WHERE expires_at < NOW()
    `)
	if err != nil {
		w.log.Error("password_resets cleanup failed", "error", err)
	} else {
		w.log.Info("password_resets cleanup", "deleted_rows", result.RowsAffected())
	}
}
