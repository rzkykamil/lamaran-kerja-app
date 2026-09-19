package web

import (
	"testing"

	lamarankerja "lamarankerja"
)

// Template hanya di-parse saat aplikasi start, jadi salah tulis di file HTML
// baru ketahuan waktu dijalankan. Test ini memajukan kegagalannya ke waktu build.
func TestSemuaTemplateBisaDiparse(t *testing.T) {
	pages, err := parsePages(lamarankerja.Assets)
	if err != nil {
		t.Fatal(err)
	}

	wajib := []string{"login.html", "dashboard.html", "lamaran_list.html",
		"lamaran_form.html", "freelance_list.html", "freelance_form.html", "pilihan.html"}
	for _, nama := range wajib {
		if _, ok := pages[nama]; !ok {
			t.Errorf("template %s tidak ikut ter-parse", nama)
		}
	}
}

func TestRupiah(t *testing.T) {
	kasus := map[int64]string{
		0:        "Rp 0",
		1000:     "Rp 1.000",
		12000000: "Rp 12.000.000",
		-500000:  "-Rp 500.000",
	}
	for input, mau := range kasus {
		if got := Rupiah(input); got != mau {
			t.Errorf("Rupiah(%d) = %q, mau %q", input, got, mau)
		}
	}
}
