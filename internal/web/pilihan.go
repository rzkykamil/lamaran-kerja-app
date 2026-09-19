package web

import (
	"net/http"

	"lamarankerja/internal/store"
)

func (a *App) pilihanList(w http.ResponseWriter, r *http.Request) {
	items, err := a.Store.ListItems(r.Context())
	if err != nil {
		a.serverError(w, r, err, "list pilihan")
		return
	}

	perKategori := map[string][]any{}
	for _, it := range items {
		perKategori[it.Category] = append(perKategori[it.Category], it)
	}

	a.render(w, r, "pilihan.html", "Pilihan Dropdown", "pilihan", map[string]any{
		"Kategori":    store.KategoriList,
		"PerKategori": perKategori,
	})
}

func (a *App) pilihanTambah(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	kategori := teks(r, "category")
	nilai := teks(r, "value")

	if nilai == "" || !kategoriValid(kategori) {
		a.flash(r, "Nilai tidak boleh kosong.", "error")
		http.Redirect(w, r, "/pilihan", http.StatusSeeOther)
		return
	}
	if err := a.Store.TambahPilihan(r.Context(), kategori, nilai); err != nil {
		a.serverError(w, r, err, "tambah pilihan")
		return
	}
	a.flash(r, "Pilihan "+nilai+" ditambahkan.", "sukses")
	http.Redirect(w, r, "/pilihan", http.StatusSeeOther)
}

func (a *App) pilihanHapus(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := a.Store.HapusPilihan(r.Context(), id); err != nil {
		a.serverError(w, r, err, "hapus pilihan")
		return
	}
	a.flash(r, "Pilihan dihapus. Data lama yang memakai nilai ini tidak ikut berubah.", "sukses")
	http.Redirect(w, r, "/pilihan", http.StatusSeeOther)
}

func kategoriValid(k string) bool {
	for _, kat := range store.KategoriList {
		if kat.Key == k {
			return true
		}
	}
	return false
}
