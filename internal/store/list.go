package store

import (
	"context"

	"lamarankerja/internal/model"
)

// Kategori dropdown beserta label yang tampil di UI.
var KategoriList = []struct {
	Key   string
	Label string
}{
	{"status_fulltime", "Status Lamaran"},
	{"sumber_fulltime", "Sumber Lowongan"},
	{"tipe_kerja", "Tipe Kerja"},
	{"prioritas", "Prioritas"},
	{"status_freelance", "Status Freelance"},
	{"sumber_freelance", "Sumber Freelance"},
	{"status_bayar", "Status Bayar"},
	{"stack", "Stack"},
}

// Pilihan adalah semua nilai dropdown, dikelompokkan per kategori.
type Pilihan map[string][]string

func (s *Store) SemuaPilihan(ctx context.Context) (Pilihan, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT category, value FROM list_items ORDER BY category, sort_order, value`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	p := Pilihan{}
	for rows.Next() {
		var cat, val string
		if err := rows.Scan(&cat, &val); err != nil {
			return nil, err
		}
		p[cat] = append(p[cat], val)
	}
	return p, rows.Err()
}

func (s *Store) ListItems(ctx context.Context) ([]model.ListItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, category, value, sort_order FROM list_items ORDER BY category, sort_order, value`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ListItem
	for rows.Next() {
		var it model.ListItem
		if err := rows.Scan(&it.ID, &it.Category, &it.Value, &it.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) TambahPilihan(ctx context.Context, category, value string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO list_items (category, value, sort_order)
		VALUES ($1, $2, COALESCE((SELECT MAX(sort_order) FROM list_items WHERE category = $1), 0) + 1)
		ON CONFLICT (category, value) DO NOTHING`, category, value)
	return err
}

func (s *Store) HapusPilihan(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM list_items WHERE id = $1`, id)
	return err
}
