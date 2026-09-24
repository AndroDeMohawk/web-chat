package repository

import (
	"github.com/AndroDeMohawk/web-chat/internal/repository/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	*db.Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, Queries: db.New(pool)}
}
func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// Pool возвращает текущий пул на случай, если потребуется выполнить транзацию вручную
func (r *Repository) Pool() *pgxpool.Pool {
	return r.pool
}
