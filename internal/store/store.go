package store

import "github.com/jackc/pgx/v5/pgxpool"

// PerHalaman adalah jumlah baris per halaman untuk daftar lamaran dan freelance.
const PerHalaman = 20

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
