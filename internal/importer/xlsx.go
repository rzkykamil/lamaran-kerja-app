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

	adaLamaran, err := importLamaran(ctx, st, f, &h)
	if err != nil {
		return h, err
	}
	adaFreelance, err := importFreelance(ctx, st, f, &h)
	if err != nil {
		return h, err
	}
	if !adaLamaran && !adaFreelance {
		return h, fmt.Errorf("file tidak punya sheet lamaran (%s) maupun Freelance",
			strings.Join(sheetLamaran, ", "))
	}
	return h, nil
}

// Sheet lamaran bisa bernama "Full-Time" (tracker Excel lama) atau "Lamaran"
// (hasil export aplikasi ini sendiri).
var sheetLamaran = []string{"Full-Time", "Lamaran"}

// importLamaran mengembalikan false (tanpa error) kalau file tidak punya
// sheet lamaran sama sekali, supaya file yang cuma berisi salah satu jenis
// data (lamaran saja atau freelance saja) tetap bisa diimpor.
func importLamaran(ctx context.Context, st *store.Store, f *excelize.File, h *Hasil) (bool, error) {
	sheet, ok := cariSheet(f, sheetLamaran...)
	if !ok {
		return false, nil
	}
	rows, err := bacaSheet(f, sheet, "Perusahaan")
	if err != nil {
		return false, err
	}

	for _, row := range rows {
		if row.teks("Perusahaan") == "" {
			continue
		}
		tglApply, err := row.tanggal("Tgl Apply")
		if err != nil {
			return false, fmt.Errorf("sheet %s baris %d kolom Tgl Apply: %w", sheet, row.nomor, err)
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
			return false, err
		}
		if duplikat {
			fmt.Printf("  dilewati (sudah ada): %s - %s\n", item.Perusahaan, item.Posisi)
			h.LamaranDilewati++
			continue
		}
		if _, err := st.CreateFulltime(ctx, item); err != nil {
			return false, fmt.Errorf("gagal menyimpan %s: %w", item.Perusahaan, err)
		}
		fmt.Printf("  masuk: %s - %s (%s)\n", item.Perusahaan, item.Posisi, item.TglApply.Format("2006-01-02"))
		h.LamaranMasuk++
	}
	return true, nil
}

// importFreelance mengembalikan false (tanpa error) kalau file tidak punya
// sheet Freelance, supaya file yang cuma berisi data lamaran tetap bisa diimpor.
func importFreelance(ctx context.Context, st *store.Store, f *excelize.File, h *Hasil) (bool, error) {
	sheet, ok := cariSheet(f, "Freelance")
	if !ok {
		return false, nil
	}
	rows, err := bacaSheet(f, sheet, "Klien")
	if err != nil {
		return false, err
	}

	for _, row := range rows {
		if row.teks("Klien") == "" {
			continue
		}
		tglLead, err := row.tanggal("Tgl Masuk Lead")
		if err != nil {
			return false, fmt.Errorf("sheet %s baris %d kolom Tgl Masuk Lead: %w", sheet, row.nomor, err)
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
			return false, err
		}
		if duplikat {
			fmt.Printf("  dilewati (sudah ada): %s - %s\n", item.Klien, item.NamaProject)
			h.ProjectDilewati++
			continue
		}
		if _, err := st.CreateFreelance(ctx, item); err != nil {
			return false, fmt.Errorf("gagal menyimpan %s: %w", item.NamaProject, err)
		}
		fmt.Printf("  masuk: %s - %s (%s)\n", item.Klien, item.NamaProject, item.TglMasukLead.Format("2006-01-02"))
		h.ProjectMasuk++
	}
	return true, nil
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

// cariSheet mengembalikan nama sheet pertama dari kandidat yang benar-benar
// ada di file (nama sheet bisa berbeda antara tracker Excel lama dan hasil
// export aplikasi ini sendiri).
func cariSheet(f *excelize.File, kandidat ...string) (string, bool) {
	ada := f.GetSheetList()
	for _, nama := range kandidat {
		for _, s := range ada {
			if s == nama {
				return nama, true
			}
		}
	}
	return "", false
}

// bacaSheet membaca sheet dan mendeteksi baris headernya secara otomatis,
// yaitu baris pertama yang punya sel bernilai kolomKunci (baris 1 di hasil
// export aplikasi, baris 4 di tracker Excel lama yang punya judul & filter di atasnya).
func bacaSheet(f *excelize.File, sheet, kolomKunci string) ([]baris, error) {
	// RawCellValue supaya tanggal terbaca sebagai serial Excel, bukan string
	// hasil format tampilan yang ambigu ("09-19-26" bisa berarti dua tanggal berbeda).
	grid, err := f.GetRows(sheet, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("sheet %s tidak terbaca: %w", sheet, err)
	}

	idxHeader := -1
	for i, row := range grid {
		for _, cell := range row {
			if strings.TrimSpace(cell) == kolomKunci {
				idxHeader = i
				break
			}
		}
		if idxHeader >= 0 {
			break
		}
	}
	if idxHeader == -1 {
		return nil, fmt.Errorf("sheet %s tidak punya baris header (kolom %q tidak ditemukan)", sheet, kolomKunci)
	}

	header := grid[idxHeader]
	var out []baris
	for i := idxHeader + 1; i < len(grid); i++ {
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
