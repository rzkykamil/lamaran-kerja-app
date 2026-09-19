package importer

import (
	"testing"

	"github.com/xuri/excelize/v2"
)

// Dua sheet menyimpan tanggal dengan cara berbeda: Full-Time pakai serial Excel,
// Freelance pakai teks ISO. Test ini menjaga keduanya tetap terbaca benar.
func TestTanggalDuaFormat(t *testing.T) {
	f, err := excelize.OpenFile("../../tracker_lamaran_kamil.xlsx")
	if err != nil {
		t.Skipf("file Excel sumber tidak tersedia: %v", err)
	}
	defer f.Close()

	kasus := []struct {
		sheet, kolom, mau string
	}{
		{"Full-Time", "Tgl Apply", "2026-09-19"},
		{"Full-Time", "Tgl Update Terakhir", "2026-09-19"},
		{"Freelance", "Tgl Masuk Lead", "2026-09-10"},
		{"Freelance", "Deadline Proposal", "2026-09-25"},
		{"Freelance", "Deadline Project", "2026-10-20"},
	}

	for _, k := range kasus {
		rows, err := bacaSheet(f, k.sheet)
		if err != nil {
			t.Fatal(err)
		}
		got, err := rows[0].tanggal(k.kolom)
		if err != nil {
			t.Errorf("%s/%s: %v", k.sheet, k.kolom, err)
			continue
		}
		if got.Format("2006-01-02") != k.mau {
			t.Errorf("%s/%s = %s, mau %s", k.sheet, k.kolom, got.Format("2006-01-02"), k.mau)
		}
	}
}

func TestNilaiNumerikTerbaca(t *testing.T) {
	f, err := excelize.OpenFile("../../tracker_lamaran_kamil.xlsx")
	if err != nil {
		t.Skipf("file Excel sumber tidak tersedia: %v", err)
	}
	defer f.Close()

	rows, err := bacaSheet(f, "Freelance")
	if err != nil {
		t.Fatal(err)
	}
	r := rows[0]

	if got := r.uang("Nilai Penawaran (Rp)"); got != 12_000_000 {
		t.Errorf("nilai penawaran = %d, mau 12000000", got)
	}
	if got := r.pecahan("Probabilitas"); got != 0.5 {
		t.Errorf("probabilitas = %v, mau 0.5", got)
	}
	if got := r.teks("Klien"); got != "Toko Berkah Jaya" {
		t.Errorf("klien = %q", got)
	}
}
