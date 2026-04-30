package usecase

import (
	"context"
	"time"
)

func (u *UseCase) StartArchiveWorker(ctx context.Context) {
	const op = "usecase.StartArchiveWorker"

	u.Log.Info("starting archive worker", "op", op)

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count, err := u.DB.ArchiveOldTasks(ctx)
			if err != nil {
				u.Log.Error("error while archiving old tasks: ", "op", op, "error", err)
			} else if count > 0 {
				u.Log.Info("archive old tasks finished", "op", op, "count", count)
			}
		case <-ctx.Done():
			u.Log.Info("context done, shutting down archive worker ", "op", op)
			return
		}
	}
}
