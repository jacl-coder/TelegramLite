package repository

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
)

type PG struct {
    Pool *pgxpool.Pool
}

func NewPG(dsn string) (*PG, error) {
    pool, err := pgxpool.New(context.Background(), dsn)
    if err != nil {
        return nil, err
    }
    return &PG{Pool: pool}, nil
}

func (p *PG) Close(ctx context.Context) {
    p.Pool.Close()
}

