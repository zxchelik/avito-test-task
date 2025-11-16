package application

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(ctx context.Context, cfg *Postgres, log *slog.Logger) (*DB, error) {
	config, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		err = fmt.Errorf("failed to create pgxpool: %s", err.Error())
		return nil, err
	}
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod

	var pool *pgxpool.Pool
	const attempts = 5

	for i := 1; i <= attempts; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			pingErr := pool.Ping(pingCtx)
			cancel()

			if pingErr == nil {
				return &DB{pool}, nil
			}

			err = fmt.Errorf("ping attempt %d/%d failed: %w", i, attempts, pingErr)
			pool.Close()
		} else {
			err = fmt.Errorf("new pool attempt %d/%d failed: %w", i, attempts, err)
		}

		if i < attempts {
			time.Sleep(time.Second * time.Duration(i))
		}
	}

	return nil, err
}

func (d *DB) Close(log slog.Logger) {
	log.Info("closing pgxpool")
	d.Pool.Close()
}
