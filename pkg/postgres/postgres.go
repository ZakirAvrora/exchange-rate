// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 5
	_defaultConnTimeout  = time.Second
)

type postgresSettings struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
	pool         *pgxpool.Pool
}

type Option func(*postgresSettings)

func MaxPoolSize(size int) Option {
	return func(c *postgresSettings) {
		c.maxPoolSize = size
	}
}

func ConnAttempts(attempts int) Option {
	return func(c *postgresSettings) {
		c.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(c *postgresSettings) {
		c.connTimeout = timeout
	}
}

func New(url string, opts ...Option) (*pgxpool.Pool, error) {
	pg := &postgresSettings{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres error: parsing configuration error: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)
	for pg.connAttempts > 0 {
		pg.pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("trying to connect to postgres, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres connection error: %w", err)
	}

	return pg.pool, nil
}
