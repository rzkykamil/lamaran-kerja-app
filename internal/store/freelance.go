package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"lamarankerja/internal/model"
)

const kolomFreelance = `id, tgl_masuk_lead, klien, nama_project, stack, sumber, scope_singkat,
	nilai_penawaran, status, probabilitas, nilai_tertimbang, deadline_proposal, tgl_mulai,
	deadline_project, status_bayar, sudah_dibayar, sisa_tagihan, next_action, catatan`

func scanFreelance(row pgx.Row) (model.Freelance, error) {
	var f model.Freelance
	err := row.Scan(&f.ID, &f.TglMasukLead, &f.Klien, &f.NamaProject, &f.Stack, &f.Sumber,
		&f.ScopeSingkat, &f.NilaiPenawaran, &f.Status, &f.Probabilitas, &f.NilaiTertimbang,
		&f.DeadlineProposal, &f.TglMulai, &f.DeadlineProject, &f.StatusBayar, &f.SudahDibayar,
		&f.SisaTagihan, &f.NextAction, &f.Catatan)
	return f, err
}

var urutanFreelance = map[string]string{
	"terbaru":  "tgl_masuk_lead DESC, id DESC",
	"terlama":  "tgl_masuk_lead ASC, id ASC",
	"nilai":    "nilai_penawaran DESC, id DESC",
	"weighted": "nilai_tertimbang DESC, id DESC",
	"sisa":     "sisa_tagihan DESC, id DESC",
	"deadline": "deadline_project ASC NULLS LAST, id DESC",
	"klien":    "klien ASC, id ASC",
}

func whereFreelance(f model.FilterFreelance) (string, []any) {
	var where []string
	var args []any

	if cari := strings.TrimSpace(f.Cari); cari != "" {
		args = append(args, "%"+cari+"%")
		n := len(args)
		where = append(where, fmt.Sprintf(
			"(klien ILIKE $%d OR nama_project ILIKE $%d OR scope_singkat ILIKE $%d OR catatan ILIKE $%d)", n, n, n, n))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if f.StatusBayar != "" {
		args = append(args, f.StatusBayar)
		where = append(where, fmt.Sprintf("status_bayar = $%d", len(args)))
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

func (s *Store) ListFreelance(ctx context.Context, f model.FilterFreelance) ([]model.Freelance, model.Paginasi, error) {
	whereClause, args := whereFreelance(f)

	var total int
	if err := s.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM freelance_projects"+whereClause, args...).Scan(&total); err != nil {
		return nil, model.Paginasi{}, err
	}
	pg := model.HitungPaginasi(f.Halaman, PerHalaman, total)

	order, ok := urutanFreelance[f.Urut]
	if !ok {
		order = urutanFreelance["terbaru"]
	}
	q := "SELECT " + kolomFreelance + " FROM freelance_projects" + whereClause + " ORDER BY " + order
	q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	argsQ := append(append([]any{}, args...), pg.PerHalaman, (pg.Halaman-1)*pg.PerHalaman)

	rows, err := s.pool.Query(ctx, q, argsQ...)
	if err != nil {
		return nil, model.Paginasi{}, err
	}
	defer rows.Close()

	var out []model.Freelance
	for rows.Next() {
		item, err := scanFreelance(rows)
		if err != nil {
			return nil, model.Paginasi{}, err
		}
		out = append(out, item)
	}
	return out, pg, rows.Err()
}

// AgregatFreelance menjumlah nilai, weighted, dan sisa tagihan dari SEMUA baris
// yang cocok dengan filter (bukan cuma satu halaman), untuk baris total di kaki tabel.
func (s *Store) AgregatFreelance(ctx context.Context, f model.FilterFreelance) (nilai, weighted, sisa int64, err error) {
	whereClause, args := whereFreelance(f)
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(nilai_penawaran),0), COALESCE(SUM(nilai_tertimbang),0), COALESCE(SUM(sisa_tagihan),0)
		 FROM freelance_projects`+whereClause, args...).Scan(&nilai, &weighted, &sisa)
	return
}

// ListFreelanceSemua mengambil SEMUA baris yang cocok filter tanpa dipotong halaman,
// dipakai untuk export supaya file yang diunduh tidak cuma berisi satu halaman.
func (s *Store) ListFreelanceSemua(ctx context.Context, f model.FilterFreelance) ([]model.Freelance, error) {
	whereClause, args := whereFreelance(f)
	order, ok := urutanFreelance[f.Urut]
	if !ok {
		order = urutanFreelance["terbaru"]
	}
	q := "SELECT " + kolomFreelance + " FROM freelance_projects" + whereClause + " ORDER BY " + order

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Freelance
	for rows.Next() {
		item, err := scanFreelance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetFreelance(ctx context.Context, id int64) (model.Freelance, error) {
	return scanFreelance(s.pool.QueryRow(ctx,
		"SELECT "+kolomFreelance+" FROM freelance_projects WHERE id = $1", id))
}

// nilai_tertimbang dan sisa_tagihan sengaja tidak ditulis: keduanya kolom GENERATED,
// PostgreSQL menolak kalau dimasukkan ke INSERT/UPDATE.
func (s *Store) CreateFreelance(ctx context.Context, f model.Freelance) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO freelance_projects
			(tgl_masuk_lead, klien, nama_project, stack, sumber, scope_singkat, nilai_penawaran,
			 status, probabilitas, deadline_proposal, tgl_mulai, deadline_project,
			 status_bayar, sudah_dibayar, next_action, catatan)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id`,
		f.TglMasukLead, f.Klien, f.NamaProject, f.Stack, f.Sumber, f.ScopeSingkat, f.NilaiPenawaran,
		f.Status, f.Probabilitas, f.DeadlineProposal, f.TglMulai, f.DeadlineProject,
		f.StatusBayar, f.SudahDibayar, f.NextAction, f.Catatan).Scan(&id)
	return id, err
}

func (s *Store) UpdateFreelance(ctx context.Context, f model.Freelance) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE freelance_projects SET
			tgl_masuk_lead = $1, klien = $2, nama_project = $3, stack = $4, sumber = $5,
			scope_singkat = $6, nilai_penawaran = $7, status = $8, probabilitas = $9,
			deadline_proposal = $10, tgl_mulai = $11, deadline_project = $12,
			status_bayar = $13, sudah_dibayar = $14, next_action = $15, catatan = $16,
			updated_at = now()
		WHERE id = $17`,
		f.TglMasukLead, f.Klien, f.NamaProject, f.Stack, f.Sumber, f.ScopeSingkat, f.NilaiPenawaran,
		f.Status, f.Probabilitas, f.DeadlineProposal, f.TglMulai, f.DeadlineProject,
		f.StatusBayar, f.SudahDibayar, f.NextAction, f.Catatan, f.ID)
	return err
}

func (s *Store) DeleteFreelance(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM freelance_projects WHERE id = $1`, id)
	return err
}

func (s *Store) FreelanceAdaDuplikat(ctx context.Context, klien, project string, tgl time.Time) (bool, error) {
	var ada bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM freelance_projects
		WHERE klien = $1 AND nama_project = $2 AND tgl_masuk_lead = $3)`,
		klien, project, tgl).Scan(&ada)
	return ada, err
}
