// Sets up and returns a new PostgreSQL connection pool
package app

import (
	"chat-message-service/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, config *config.Config) (*pgxpool.Pool, error) {
	log.Printf("Connecting to DB at %s", config.DbUrl)
	cfg, err := pgxpool.ParseConfig(config.DbUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DB config: %v", err)
	}

	if config.DbMaxConns > 0 {
		cfg.MaxConns = config.DbMaxConns
	}
	if config.DbMaxIdleTimeSec > 0 {
		cfg.MaxConnIdleTime = time.Duration(config.DbMaxIdleTimeSec) * time.Second
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Printf("Error pinging DB at %s: %s", config.DbUrl, err)
		return nil, fmt.Errorf("unable to ping DB: %v", err)
	}

	return pool, nil
}
