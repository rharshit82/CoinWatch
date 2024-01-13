package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func NewPostresDB(ctx context.Context, addr string) (Querier, *pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, addr)
	if err != nil {
		return nil, nil, err
	}

	db := New(conn)
	return db, conn, nil
}
