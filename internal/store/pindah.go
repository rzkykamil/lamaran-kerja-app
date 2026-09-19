package store

import "context"

// PindahKeFreelance memindahkan satu lamaran full-time menjadi project freelance:
// insert ke freelance_projects lalu hapus dari fulltime_applications, dalam satu
// transaksi supaya datanya tidak hilang kalau salah satu langkah gagal.
func (s *Store) PindahKeFreelance(ctx context.Context, id int64) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	f, err := scanFulltime(tx.QueryRow(ctx,
		"SELECT "+kolomFulltime+" FROM fulltime_applications WHERE id = $1", id))
	if err != nil {
		return 0, err
	}

	var newID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO freelance_projects
			(tgl_masuk_lead, klien, nama_project, stack, sumber, scope_singkat, nilai_penawaran,
			 next_action, catatan)
		VALUES ($1,$2,$3,'',$4,'',$5,$6,$7)
		RETURNING id`,
		f.TglApply, f.Perusahaan, f.Posisi, f.Sumber, f.EkspektasiGaji, f.NextAction, f.Catatan,
	).Scan(&newID)
	if err != nil {
		return 0, err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM fulltime_applications WHERE id = $1", id); err != nil {
		return 0, err
	}

	return newID, tx.Commit(ctx)
}

// PindahKeLamaran adalah kebalikan dari PindahKeFreelance: memindahkan satu
// project freelance menjadi lamaran full-time.
func (s *Store) PindahKeLamaran(ctx context.Context, id int64) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	f, err := scanFreelance(tx.QueryRow(ctx,
		"SELECT "+kolomFreelance+" FROM freelance_projects WHERE id = $1", id))
	if err != nil {
		return 0, err
	}

	var newID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO fulltime_applications
			(tgl_apply, perusahaan, posisi, sumber, ekspektasi_gaji, next_action, catatan)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id`,
		f.TglMasukLead, f.Klien, f.NamaProject, f.Sumber, f.NilaiPenawaran, f.NextAction, f.Catatan,
	).Scan(&newID)
	if err != nil {
		return 0, err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM freelance_projects WHERE id = $1", id); err != nil {
		return 0, err
	}

	return newID, tx.Commit(ctx)
}
