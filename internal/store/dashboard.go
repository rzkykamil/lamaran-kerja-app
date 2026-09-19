package store

import (
	"context"

	"lamarankerja/internal/model"
)

// Label kartu target, mengikuti sheet Dashboard di Excel aslinya.
var labelTarget = []struct{ Key, Label, Catatan string }{
	{"perusahaan_dilamar", "Perusahaan dilamar", "Total lamaran terkirim"},
	{"masuk_interview", "Masuk tahap interview", "Interview HR ke atas"},
	{"lolos_interview2", "Lolos ke interview ke-2", "Interview User / Final / Test / Offer"},
	{"offer_diterima", "Offer diterima", "Status lamaran sudah Diterima"},
	{"freelance_closing", "Freelance closing", "Project berstatus Deal / DP ke atas"},
}

func (s *Store) Dashboard(ctx context.Context) (model.Dashboard, error) {
	var d model.Dashboard

	var lanjut int
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('Interview HR','Interview User','Interview Final',
			                                  'Technical Test','Offer','Diterima')),
			COUNT(*) FILTER (WHERE status IN ('Interview User','Interview Final','Technical Test',
			                                  'Offer','Diterima')),
			COUNT(*) FILTER (WHERE status = 'Diterima'),
			COUNT(*) FILTER (WHERE status IN ('Screening HR','Interview HR','Interview User',
			                                  'Interview Final','Technical Test','Offer','Diterima')),
			COUNT(*) FILTER (WHERE tgl_update_terakhir < CURRENT_DATE - 14
			                   AND status NOT IN ('Diterima','Ditolak','Withdraw','Ghosting'))
		FROM fulltime_applications`).Scan(
		&d.TotalLamaran, &d.MasukInterview, &d.LolosInterview2, &d.OfferDiterima, &lanjut, &d.Nganggur)
	if err != nil {
		return d, err
	}
	if d.TotalLamaran > 0 {
		d.ResponseRate = float64(lanjut) / float64(d.TotalLamaran)
	}

	err = s.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(nilai_penawaran), 0),
			COALESCE(SUM(nilai_tertimbang), 0),
			COALESCE(SUM(nilai_penawaran) FILTER (
				WHERE status IN ('Deal / DP','Dikerjakan','Review Klien','Selesai')), 0),
			COALESCE(SUM(sudah_dibayar), 0),
			COALESCE(SUM(sisa_tagihan) FILTER (WHERE status <> 'Batal'), 0),
			COUNT(*) FILTER (WHERE status <> 'Lead'),
			COUNT(*) FILTER (WHERE status IN ('Deal / DP','Dikerjakan','Review Klien','Selesai'))
		FROM freelance_projects`).Scan(
		&d.TotalPipeline, &d.NilaiTertimbang, &d.NilaiClosing, &d.UangMasuk, &d.SisaTagihan,
		&d.ProposalKeluar, &d.FreelanceDeal)
	if err != nil {
		return d, err
	}
	if d.ProposalKeluar > 0 {
		d.WinRate = float64(d.FreelanceDeal) / float64(d.ProposalKeluar)
	}

	if d.FunnelFulltime, err = s.funnel(ctx, "status_fulltime", "fulltime_applications"); err != nil {
		return d, err
	}
	if d.FunnelFreelance, err = s.funnel(ctx, "status_freelance", "freelance_projects"); err != nil {
		return d, err
	}

	targets, err := s.Targets(ctx)
	if err != nil {
		return d, err
	}
	aktual := map[string]int{
		"perusahaan_dilamar": d.TotalLamaran,
		"masuk_interview":    d.MasukInterview,
		"lolos_interview2":   d.LolosInterview2,
		"offer_diterima":     d.OfferDiterima,
		"freelance_closing":  d.FreelanceDeal,
	}
	for _, t := range labelTarget {
		row := model.TargetRow{
			Key:     t.Key,
			Label:   t.Label,
			Catatan: t.Catatan,
			Target:  targets[t.Key],
			Aktual:  aktual[t.Key],
		}
		if row.Target > 0 {
			row.Progress = min(row.Aktual*100/row.Target, 100)
		}
		d.Targets = append(d.Targets, row)
	}

	if d.PerluFollowUp, err = s.PerluFollowUp(ctx); err != nil {
		return d, err
	}
	return d, nil
}

// funnel mengambil semua nilai status dari list_items supaya status yang jumlahnya 0 tetap muncul.
func (s *Store) funnel(ctx context.Context, kategori, tabel string) ([]model.FunnelRow, error) {
	q := `SELECT li.value, COUNT(t.id)
		FROM list_items li
		LEFT JOIN ` + tabel + ` t ON t.status = li.value
		WHERE li.category = $1
		GROUP BY li.value, li.sort_order
		ORDER BY li.sort_order`

	rows, err := s.pool.Query(ctx, q, kategori)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.FunnelRow
	for rows.Next() {
		var r model.FunnelRow
		if err := rows.Scan(&r.Label, &r.Jumlah); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// PerluFollowUp: lamaran yang tanggal follow-upnya sudah lewat, atau nganggur lebih dari 14 hari.
func (s *Store) PerluFollowUp(ctx context.Context) ([]model.Fulltime, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+kolomFulltime+` FROM fulltime_applications
		WHERE status NOT IN ('Diterima','Ditolak','Withdraw','Ghosting')
		  AND (tgl_follow_up < CURRENT_DATE OR tgl_update_terakhir < CURRENT_DATE - 14)
		ORDER BY COALESCE(tgl_follow_up, tgl_update_terakhir) ASC`)
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

func (s *Store) Targets(ctx context.Context) (map[string]int, error) {
	rows, err := s.pool.Query(ctx, `SELECT metric_key, target_value FROM dashboard_targets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var k string
		var v int
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func (s *Store) SimpanTarget(ctx context.Context, key string, value int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO dashboard_targets (metric_key, target_value) VALUES ($1, $2)
		ON CONFLICT (metric_key) DO UPDATE SET target_value = EXCLUDED.target_value`, key, value)
	return err
}
