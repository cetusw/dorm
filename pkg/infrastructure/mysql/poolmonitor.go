package mysql

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// MonitorPool samples counters without performing SQL and returns on cancellation.
func MonitorPool(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	previous := db.Stats()
	logPoolStats(ctx, previous, nil, true)
	lastInfo := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
			stats := db.Stats()
			info := time.Since(lastInfo) >= 5*time.Minute
			logPoolStats(ctx, stats, &previous, info)
			if info {
				lastInfo = time.Now()
			}
			previous = stats
		}
	}
}

func logPoolStats(ctx context.Context, stats sql.DBStats, previous *sql.DBStats, info bool) {
	var waitCount int64
	var waitDuration time.Duration
	deltaAvailable := previous != nil && stats.WaitCount >= previous.WaitCount && stats.WaitDuration >= previous.WaitDuration
	if deltaAvailable {
		waitCount = stats.WaitCount - previous.WaitCount
		waitDuration = stats.WaitDuration - previous.WaitDuration
	}
	saturated := (stats.MaxOpenConnections == 25 && stats.InUse >= 20) || waitDuration > 500*time.Millisecond
	if !info && !saturated {
		return
	}
	level, event := slog.LevelInfo, "db.pool_stats"
	if saturated {
		level, event = slog.LevelWarn, "db.pool_saturation"
	}
	fields := []any{
		"event", event, "max_open", stats.MaxOpenConnections, "open", stats.OpenConnections,
		"in_use", stats.InUse, "idle", stats.Idle,
		"wait_count_total", stats.WaitCount, "wait_duration_ms_total", stats.WaitDuration.Milliseconds(),
		"max_idle_closed", stats.MaxIdleClosed, "max_idle_time_closed", stats.MaxIdleTimeClosed,
		"max_lifetime_closed", stats.MaxLifetimeClosed, "delta_available", deltaAvailable,
	}
	if deltaAvailable {
		fields = append(fields, "wait_count_delta", waitCount, "wait_duration_ms_delta", waitDuration.Milliseconds())
	}
	slog.Log(ctx, level, event, fields...)
}
