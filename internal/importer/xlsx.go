// Package importer memindahkan isi tracker Excel lama ke database.
package importer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"lamarankerja/internal/model"
	"lamarankerja/internal/store"
)

// barisHeader adalah baris ke-4 di kedua sheet; data mulai baris ke-5.
const barisHeader = 4

type Hasil struct {
	LamaranMasuk    int
	LamaranDilewati int
	ProjectMasuk    int
	ProjectDilewati int
}

func Import(ctx context.Context, st *store.Store, path string) (Hasil, error) {
	var h Hasil

	f, err := excelize.OpenFile(path)
	if err != nil {
		return h, fmt.Errorf("gagal membuka %s (pastikan filenya tidak sedang dibuka di Excel): %w", path, err)
	}
	defer f.Close()

	if err := importLamaran(ctx, st, f, &h); err != nil {
		return h, err
	}
	if err := importFreelance(ctx, st, f, &h); err != nil {
		return h, err
	}
	return h, nil
}

func importLamaran(ctx context.Context, st *store.Store, f *excelize.File, h *Hasil) error {
	const sheet = "Full-Time"
	rows, err := bacaSheet(f, sheet)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if row.teks("Perusahaan") == "" {
			continue
		}
		tglApply, err := row.tanggal("Tgl Apply")
		if err != nil {
			return fmt.Errorf("sheet %s baris %d kolom Tgl Apply: %w", sheet, row.nomor, err)
		}
		tglUpdate, err := row.tanggal("Tgl Update Terakhir")
		if err != nil || tglUpdate.IsZero() {
			tglUpdate = tglApply
		}

		item := model.Fulltime{
			TglApply:          tglApply,
			Perusahaan:        row.teks("Perusahaan"),
			Posisi:            row.teks("Posisi"),
			Lokasi:            row.teks("Lokasi"),
			TipeKerja:         row.teks("Tipe Kerja"),
			Sumber:            row.teks("Sumber"),
			LinkLowongan:      row.teks("Link Lowongan"),
			EkspektasiGaji:    row.uang("Ekspektasi Gaji (Rp)"),
			Prioritas:         nilaiAtau(row.teks("Prioritas"), "Sedang"),
			Status:            nilaiAtau(row.teks("Status"), "Apply"),
			TahapKe:           row.bilangan("Tahap Ke-"),
			TglUpdateTerakhir: tglUpdate,
			NextAction:        row.teks("Next Action"),
			Kontak:            row.teks("Kontak (nama / email)"),
			Catatan:           row.teks("Catatan"),
		}
		if t, err := row.tanggal("Tgl Follow-up"); err == nil && !t.IsZero() {
			item.TglFollowUp = &t
		}

		duplikat, err := st.FulltimeAdaDuplikat(ctx, item.Perusahaan, item.Posisi, item.TglApply)
		if err != nil {
			return err
		}
		if duplikat {
			fmt.Printf("  dilewati (sudah ada): %s - %s\n", item.Perusahaan, item.Posisi)
			h.LamaranDilewati++
			continue
		}
		if _, err := st.CreateFulltime(ctx, item); err != nil {
			return fmt.Errorf("gagal menyimpan %s: %w", item.Perusahaan, err)
		}
		fmt.Printf("  masuk: %s - %s (%s)\n", item.Perusahaan, item.Posisi, item.TglApply.Format("2006-01-02"))
		h.LamaranMasuk++
	}
	return nil
}

func importFreelance(ctx context.Context, st *store.Store, f *excelize.File, h *Hasil) error {
	const sheet = "Freelance"
	rows, err := bacaSheet(f, sheet)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if row.teks("Klien") == "" {
			continue
		}
		tglLead, err := row.tanggal("Tgl Masuk Lead")
		if err != nil {
			return fmt.Errorf("sheet %s baris %d kolom Tgl Masuk Lead: %w", sheet, row.nomor, err)
		}

		item := model.Freelance{
			TglMasukLead:   tglLead,
			Klien:          row.teks("Klien"),
			NamaProject:    row.teks("Nama Project"),
			Stack:          row.teks("Stack"),
			Sumber:         row.teks("Sumber"),
			ScopeSingkat:   row.teks("Scope Singkat"),
			NilaiPenawaran: row.uang("Nilai Penawaran (Rp)"),
			Status:         nilaiAtau(row.teks("Status"), "Lead"),
			Probabilitas:   row.pecahan("Probabilitas"),
			StatusBayar:    nilaiAtau(row.teks("Status Bayar"), "Belum Ditagih"),
			SudahDibayar:   row.uang("Sudah Dibayar (Rp)"),
			NextAction:     row.teks("Next Action"),
			Catatan:        row.teks("Catatan"),
		}
		for kolom, tujuan := range map[string]**time.Time{
			"Deadline Proposal": &item.DeadlineProposal,
			"Tgl Mulai":         &item.TglMulai,
			"Deadline Project":  &item.DeadlineProject,
		} {
			if t, err := row.tanggal(kolom); err == nil && !t.IsZero() {
				tt := t
				*tujuan = &tt
			}
		}

		duplikat, err := st.FreelanceAdaDuplikat(ctx, item.Klien, item.NamaProject, item.TglMasukLead)
		if err != nil {
			return err
		}
		if duplikat {
			fmt.Printf("  dilewati (sudah ada): %s - %s\n", item.Klien, item.NamaProject)
			h.ProjectDilewati++
			continue
		}
		if _, err := st.CreateFreelance(ctx, item); err != nil {
			return fmt.Errorf("gagal menyimpan %s: %w", item.NamaProject, err)
		}
		fmt.Printf("  masuk: %s - %s (%s)\n", item.Klien, item.NamaProject, item.TglMasukLead.Format("2006-01-02"))
		h.ProjectMasuk++
	}
	return nil
}

// baris memetakan nama header ke nilai mentah selnya.
type baris struct {
	nomor int
	sel   map[string]string
}

func (b baris) teks(kolom string) string { return strings.TrimSpace(b.sel[kolom]) }

func (b baris) bilangan(kolom string) int {
	n, _ := strconv.Atoi(b.teks(kolom))
	return n
}

func (b baris) uang(kolom string) int64 {
	f, err := strconv.ParseFloat(b.teks(kolom), 64)
	if err != nil {
		return 0
	}
	return int64(f)
}

func (b baris) pecahan(kolom string) float64 {
	f, err := strconv.ParseFloat(b.teks(kolom), 64)
	if err != nil {
		return 0
	}
	// Sebagian orang menulis 50 dan bukan 0.5.
	if f > 1 {
		f = f / 100
	}
	return f
}

// tanggal menangani dua bentuk yang dipakai di file aslinya sekaligus:
// sheet Full-Time menyimpan serial Excel ("46284"), sheet Freelance menyimpan teks "2026-09-10".
func (b baris) tanggal(kolom string) (time.Time, error) {
	raw := b.teks(kolom)
	if raw == "" {
		return time.Time{}, nil
	}
	if serial, err := strconv.ParseFloat(raw, 64); err == nil {
		t, err := excelize.ExcelDateToTime(serial, false)
		if err != nil {
			return time.Time{}, err
		}
		return t.Truncate(24 * time.Hour), nil
	}
	for _, layout := range []string{"2006-01-02", "02/01/2006", "01-02-06", "2006/01/02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("format tanggal %q tidak dikenali", raw)
}

func bacaSheet(f *excelize.File, sheet string) ([]baris, error) {
	// RawCellValue supaya tanggal terbaca sebagai serial Excel, bukan string
	// hasil format tampilan yang ambigu ("09-19-26" bisa berarti dua tanggal berbeda).
	grid, err := f.GetRows(sheet, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("sheet %s tidak terbaca: %w", sheet, err)
	}
	if len(grid) < barisHeader {
		return nil, fmt.Errorf("sheet %s tidak punya baris header di baris %d", sheet, barisHeader)
	}

	header := grid[barisHeader-1]
	var out []baris
	for i := barisHeader; i < len(grid); i++ {
		row := baris{nomor: i + 1, sel: map[string]string{}}
		for c, nama := range header {
			nama = strings.TrimSpace(nama)
			if nama == "" || c >= len(grid[i]) {
				continue
			}
			row.sel[nama] = grid[i][c]
		}
		out = append(out, row)
	}
	return out, nil
}

func nilaiAtau(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
