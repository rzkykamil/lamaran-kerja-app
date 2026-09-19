package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"lamarankerja/internal/model"
)

const kolomFulltime = `id, tgl_apply, perusahaan, posisi, lokasi, tipe_kerja, sumber, link_lowongan,
	ekspektasi_gaji, prioritas, status, tahap_ke, tgl_update_terakhir,
	(CURRENT_DATE - tgl_update_terakhir) AS umur_hari,
	next_action, tgl_follow_up, kontak, catatan`

func scanFulltime(row pgx.Row) (model.Fulltime, error) {
	var f model.Fulltime
	err := row.Scan(&f.ID, &f.TglApply, &f.Perusahaan, &f.Posisi, &f.Lokasi, &f.TipeKerja,
		&f.Sumber, &f.LinkLowongan, &f.EkspektasiGaji, &f.Prioritas, &f.Status, &f.TahapKe,
		&f.TglUpdateTerakhir, &f.UmurHari, &f.NextAction, &f.TglFollowUp, &f.Kontak, &f.Catatan)
	return f, err
}

// urutanFulltime memetakan pilihan urutan dari UI ke klausa ORDER BY.
// Hanya nilai dari map ini yang boleh masuk ke SQL, supaya tidak bisa disuntik.
var urutanFulltime = map[string]string{
	"terbaru":   "tgl_apply DESC, id DESC",
	"terlama":   "tgl_apply ASC, id ASC",
	"umur":      "umur_hari DESC, id DESC",
	"gaji":      "ekspektasi_gaji DESC, id DESC",
	"nama":      "perusahaan ASC, id ASC",
	"prioritas": "CASE prioritas WHEN 'Tinggi' THEN 1 WHEN 'Sedang' THEN 2 ELSE 3 END, id DESC",
}

func whereFulltime(f model.FilterFulltime) (string, []any) {
	var where []string
	var args []any

	if cari := strings.TrimSpace(f.Cari); cari != "" {
		args = append(args, "%"+cari+"%")
		n := len(args)
		where = append(where, fmt.Sprintf(
			"(perusahaan ILIKE $%d OR posisi ILIKE $%d OR lokasi ILIKE $%d OR catatan ILIKE $%d)", n, n, n, n))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if f.Prioritas != "" {
		args = append(args, f.Prioritas)
		where = append(where, fmt.Sprintf("prioritas = $%d", len(args)))
	}
	if f.Sumber != "" {
		args = append(args, f.Sumber)
		where = append(where, fmt.Sprintf("sumber = $%d", len(args)))
	}

	if len(where) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(where, " AND "), args
}

func (s *Store) ListFulltime(ctx context.Context, f model.FilterFulltime) ([]model.Fulltime, model.Paginasi, error) {
	whereClause, args := whereFulltime(f)

	var total int
	if err := s.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM fulltime_applications"+whereClause, args...).Scan(&total); err != nil {
		return nil, model.Paginasi{}, err
	}
	pg := model.HitungPaginasi(f.Halaman, PerHalaman, total)

	order, ok := urutanFulltime[f.Urut]
	if !ok {
		order = urutanFulltime["terbaru"]
	}
	q := "SELECT " + kolomFulltime + " FROM fulltime_applications" + whereClause + " ORDER BY " + order
	q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	argsQ := append(append([]any{}, args...), pg.PerHalaman, (pg.Halaman-1)*pg.PerHalaman)

	rows, err := s.pool.Query(ctx, q, argsQ...)
	if err != nil {
		return nil, model.Paginasi{}, err
	}
	defer rows.Close()

	var out []model.Fulltime
	for rows.Next() {
		item, err := scanFulltime(rows)
		if err != nil {
			return nil, model.Paginasi{}, err
		}
		out = append(out, item)
	}
	return out, pg, rows.Err()
}

// ListFulltimeSemua mengambil SEMUA baris yang cocok filter tanpa dipotong halaman,
// dipakai untuk export supaya file yang diunduh tidak cuma berisi satu halaman.
func (s *Store) ListFulltimeSemua(ctx context.Context, f model.FilterFulltime) ([]model.Fulltime, error) {
	whereClause, args := whereFulltime(f)
	order, ok := urutanFulltime[f.Urut]
	if !ok {
		order = urutanFulltime["terbaru"]
	}
	q := "SELECT " + kolomFulltime + " FROM fulltime_applications" + whereClause + " ORDER BY " + order

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Fulltime
	for rows.Next() {
		item, err := scanFulltime(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetFulltime(ctx context.Context, id int64) (model.Fulltime, error) {
	return scanFulltime(s.pool.QueryRow(ctx,
		"SELECT "+kolomFulltime+" FROM fulltime_applications WHERE id = $1", id))
}

func (s *Store) CreateFulltime(ctx context.Context, f model.Fulltime) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO fulltime_applications
			(tgl_apply, perusahaan, posisi, lokasi, tipe_kerja, sumber, link_lowongan,
			 ekspektasi_gaji, prioritas, status, tahap_ke, tgl_update_terakhir,
			 next_action, tgl_follow_up, kontak, catatan)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id`,
		f.TglApply, f.Perusahaan, f.Posisi, f.Lokasi, f.TipeKerja, f.Sumber, f.LinkLowongan,
		f.EkspektasiGaji, f.Prioritas, f.Status, f.TahapKe, f.TglUpdateTerakhir,
		f.NextAction, f.TglFollowUp, f.Kontak, f.Catatan).Scan(&id)
	return id, err
}

func (s *Store) UpdateFulltime(ctx context.Context, f model.Fulltime) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE fulltime_applications SET
			tgl_apply = $1, perusahaan = $2, posisi = $3, lokasi = $4, tipe_kerja = $5,
			sumber = $6, link_lowongan = $7, ekspektasi_gaji = $8, prioritas = $9,
			status = $10, tahap_ke = $11, tgl_update_terakhir = $12, next_action = $13,
			tgl_follow_up = $14, kontak = $15, catatan = $16, updated_at = now()
		WHERE id = $17`,
		f.TglApply, f.Perusahaan, f.Posisi, f.Lokasi, f.TipeKerja, f.Sumber, f.LinkLowongan,
		f.EkspektasiGaji, f.Prioritas, f.Status, f.TahapKe, f.TglUpdateTerakhir,
		f.NextAction, f.TglFollowUp, f.Kontak, f.Catatan, f.ID)
	return err
}

func (s *Store) DeleteFulltime(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM fulltime_applications WHERE id = $1`, id)
	return err
}

// FulltimeAdaDuplikat dipakai importer supaya aman dijalankan berulang kali.
func (s *Store) FulltimeAdaDuplikat(ctx context.Context, perusahaan, posisi string, tgl time.Time) (bool, error) {
	var ada bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM fulltime_applications
		WHERE perusahaan = $1 AND posisi = $2 AND tgl_apply = $3)`,
		perusahaan, posisi, tgl).Scan(&ada)
	return ada, err
}
