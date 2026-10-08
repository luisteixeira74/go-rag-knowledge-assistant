package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go/pgx"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao parsear DSN do postgres: %w", err)
	}

	// Configuração do pool de conexões
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	// Registra o tipo do pgvector para que o pgx entenda []float32 como vector
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgvector.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pool de conexões: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("falha ao verificar conexão com postgres: %w", err)
	}

	return pool, nil
}
