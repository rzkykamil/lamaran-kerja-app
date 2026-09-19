package web

import (
	"io"
	"testing"
	"time"

	lamarankerja "lamarankerja"
	"lamarankerja/internal/model"
	"lamarankerja/internal/store"
)

func contohPilihan() store.Pilihan {
	return store.Pilihan{
		"status_fulltime":  {"Wishlist", "Apply", "Interview HR", "Diterima"},
		"sumber_fulltime":  {"LinkedIn", "JobStreet"},
		"tipe_kerja":       {"WFO", "Hybrid"},
		"prioritas":        {"Tinggi", "Sedang", "Rendah"},
		"status_freelance": {"Lead", "Proposal Terkirim", "Selesai"},
		"sumber_freelance": {"Referral", "Upwork"},
		"status_bayar":     {"Belum Ditagih", "Lunas"},
		"stack":            {"Golang", "Next.js"},
	}
}

func contohLamaran() model.Fulltime {
	kemarin := time.Now().AddDate(0, 0, -1)
	return model.Fulltime{
		ID: 1, TglApply: time.Now(), Perusahaan: "PT Contoh", Posisi: "Backend Developer",
		Lokasi: "Jakarta", TipeKerja: "WFO", Sumber: "JobStreet", LinkLowongan: "https://contoh.id",
		EkspektasiGaji: 8_000_000, Prioritas: "Sedang", Status: "Apply", TahapKe: 1,
		TglUpdateTerakhir: time.Now(), UmurHari: 20, NextAction: "Follow up HR",
		TglFollowUp: &kemarin, Kontak: "Bu Sari", Catatan: "Catatan uji",
	}
}

func contohProject() model.Freelance {
	besok := time.Now().AddDate(0, 0, 1)
	return model.Freelance{
		ID: 1, TglMasukLead: time.Now(), Klien: "Toko Berkah", NamaProject: "Company profile",
		Stack: "Next.js", Sumber: "Referral", ScopeSingkat: "Landing page 5 halaman",
		NilaiPenawaran: 12_000_000, Status: "Proposal Terkirim", Probabilitas: 0.5,
		NilaiTertimbang: 6_000_000, DeadlineProject: &besok, StatusBayar: "Belum Ditagih",
		SisaTagihan: 12_000_000, NextAction: "Follow up H+3",
	}
}

// Kesalahan seperti salah nama field atau memanggil index pada map kosong baru
// muncul saat template dieksekusi, bukan saat di-parse. Test ini menjalankan
// setiap halaman dengan data yang mirip aslinya.
func TestSemuaHalamanBisaDirender(t *testing.T) {
	pages, err := parsePages(lamarankerja.Assets)
	if err != nil {
		t.Fatal(err)
	}

	lamaran := contohLamaran()
	project := contohProject()
	pilihan := contohPilihan()

	dash := model.Dashboard{
		Targets: []model.TargetRow{
			{Key: "perusahaan_dilamar", Label: "Perusahaan dilamar", Target: 5, Aktual: 5, Progress: 100},
			{Key: "offer_diterima", Label: "Offer diterima", Target: 1, Aktual: 0, Progress: 0},
		},
		FunnelFulltime:  []model.FunnelRow{{Label: "Apply", Jumlah: 5}, {Label: "Offer", Jumlah: 0}},
		FunnelFreelance: []model.FunnelRow{{Label: "Lead", Jumlah: 0}, {Label: "Proposal Terkirim", Jumlah: 1}},
		TotalLamaran:    5,
		ResponseRate:    0.2,
		TotalPipeline:   12_000_000,
		NilaiTertimbang: 6_000_000,
		SisaTagihan:     12_000_000,
		ProposalKeluar:  1,
		PerluFollowUp:   []model.Fulltime{lamaran},
	}

	kasus := []struct {
		page string
		blok string
		data any
	}{
		{"login.html", "login", map[string]any{"Error": "Username atau password salah.", "Username": "admin"}},
		{"dashboard.html", "layout", halaman{Judul: "Dashboard", Aktif: "dashboard",
			Data: map[string]any{"D": dash, "MaxFunnel": 5, "MaxPipeline": 1, "Username": "admin"}}},
		{"lamaran_list.html", "layout", halaman{Judul: "Lamaran", Aktif: "lamaran",
			Data: map[string]any{"Items": []model.Fulltime{lamaran},
				"Filter": model.FilterFulltime{Cari: "contoh"}, "Pilihan": pilihan,
				"Paginasi": model.HitungPaginasi(1, store.PerHalaman, 1)}}},
		{"lamaran_form.html", "layout", halaman{Judul: "Tambah", Aktif: "lamaran",
			Data: map[string]any{"Item": lamaran, "Pilihan": pilihan, "Baru": false, "Error": "Posisi wajib diisi."}}},
		{"freelance_list.html", "layout", halaman{Judul: "Freelance", Aktif: "freelance",
			Data: map[string]any{"Items": []model.Freelance{project},
				"Filter": model.FilterFreelance{}, "Pilihan": pilihan,
				"Paginasi":      model.HitungPaginasi(1, store.PerHalaman, 1),
				"TotalNilai":    int64(12_000_000), "TotalWeighted": int64(6_000_000),
				"TotalSisa": int64(12_000_000)}}},
		{"freelance_form.html", "layout", halaman{Judul: "Tambah", Aktif: "freelance",
			Data: map[string]any{"Item": project, "Pilihan": pilihan, "Baru": true}}},
		{"impor.html", "layout", halaman{Judul: "Impor Data", Aktif: "impor",
			Data: map[string]any{"Error": "Pilih file Excel (.xlsx) terlebih dahulu."}}},
		{"pilihan.html", "layout", halaman{Judul: "Pilihan", Aktif: "pilihan",
			Data: map[string]any{"Kategori": store.KategoriList, "PerKategori": map[string][]any{
				"status_fulltime": {model.ListItem{ID: 1, Category: "status_fulltime", Value: "Apply"}},
			}}}},
	}

	for _, k := range kasus {
		t.Run(k.page, func(t *testing.T) {
			if err := pages[k.page].ExecuteTemplate(io.Discard, k.blok, k.data); err != nil {
				t.Errorf("render %s gagal: %v", k.page, err)
			}
		})
	}
}

// Daftar kosong harus menampilkan empty state, bukan panic.
func TestHalamanDaftarKosong(t *testing.T) {
	pages, err := parsePages(lamarankerja.Assets)
	if err != nil {
		t.Fatal(err)
	}
	pilihan := contohPilihan()

	err = pages["lamaran_list.html"].ExecuteTemplate(io.Discard, "layout", halaman{
		Data: map[string]any{"Items": []model.Fulltime{},
			"Filter": model.FilterFulltime{}, "Pilihan": pilihan,
			"Paginasi": model.HitungPaginasi(1, store.PerHalaman, 0)}})
	if err != nil {
		t.Errorf("lamaran_list kosong: %v", err)
	}

	err = pages["freelance_list.html"].ExecuteTemplate(io.Discard, "layout", halaman{
		Data: map[string]any{"Items": []model.Freelance{}, "Filter": model.FilterFreelance{},
			"Pilihan":       pilihan,
			"Paginasi":      model.HitungPaginasi(1, store.PerHalaman, 0),
			"TotalNilai":    int64(0), "TotalWeighted": int64(0), "TotalSisa": int64(0)}})
	if err != nil {
		t.Errorf("freelance_list kosong: %v", err)
	}
}
