package web

import (
	"net/http"
	"strconv"
	"time"

	"lamarankerja/internal/model"
)

func filterLamaranDariURL(r *http.Request) model.FilterFulltime {
	q := r.URL.Query()
	halaman, _ := strconv.Atoi(q.Get("halaman"))
	return model.FilterFulltime{
		Cari:      q.Get("cari"),
		Status:    q.Get("status"),
		Prioritas: q.Get("prioritas"),
		Sumber:    q.Get("sumber"),
		Urut:      q.Get("urut"),
		Halaman:   halaman,
	}
}

func (a *App) lamaranList(w http.ResponseWriter, r *http.Request) {
	filter := filterLamaranDariURL(r)
	items, paginasi, err := a.Store.ListFulltime(r.Context(), filter)
	if err != nil {
		a.serverError(w, r, err, "list lamaran")
		return
	}

	data := map[string]any{"Items": items, "Filter": filter, "Paginasi": paginasi}

	// Request dari HTMX (kotak cari / filter) cuma butuh tabelnya saja.
	if r.Header.Get("HX-Request") == "true" {
		a.renderPartial(w, "lamaran_list.html", "tabel-lamaran", data)
		return
	}

	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	data["Pilihan"] = pilihan
	a.render(w, r, "lamaran_list.html", "Lamaran Kerja", "lamaran", data)
}

func (a *App) lamaranBaru(w http.ResponseWriter, r *http.Request) {
	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	hariIni := time.Now().Truncate(24 * time.Hour)
	item := model.Fulltime{
		TglApply:          hariIni,
		TglUpdateTerakhir: hariIni,
		Status:            "Apply",
		Prioritas:         "Sedang",
		TahapKe:           1,
	}
	a.render(w, r, "lamaran_form.html", "Tambah Lamaran", "lamaran", map[string]any{
		"Item": item, "Pilihan": pilihan, "Baru": true,
	})
}

func (a *App) lamaranEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	item, err := a.Store.GetFulltime(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	a.render(w, r, "lamaran_form.html", "Edit Lamaran", "lamaran", map[string]any{
		"Item": item, "Pilihan": pilihan, "Baru": false,
	})
}

func (a *App) bacaFormLamaran(r *http.Request) (model.Fulltime, string) {
	f := model.Fulltime{
		TglApply:          tanggalWajib(r, "tgl_apply"),
		Perusahaan:        teks(r, "perusahaan"),
		Posisi:            teks(r, "posisi"),
		Lokasi:            teks(r, "lokasi"),
		TipeKerja:         teks(r, "tipe_kerja"),
		Sumber:            teks(r, "sumber"),
		LinkLowongan:      teks(r, "link_lowongan"),
		EkspektasiGaji:    uang(r, "ekspektasi_gaji"),
		Prioritas:         teks(r, "prioritas"),
		Status:            teks(r, "status"),
		TahapKe:           bilangan(r, "tahap_ke"),
		TglUpdateTerakhir: tanggalWajib(r, "tgl_update_terakhir"),
		NextAction:        teks(r, "next_action"),
		TglFollowUp:       tanggalOpsional(r, "tgl_follow_up"),
		Kontak:            teks(r, "kontak"),
		Catatan:           teks(r, "catatan"),
	}
	switch {
	case f.Perusahaan == "":
		return f, "Nama perusahaan wajib diisi."
	case f.Posisi == "":
		return f, "Posisi yang dilamar wajib diisi."
	}
	return f, ""
}

func (a *App) lamaranCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	item, pesanError := a.bacaFormLamaran(r)
	if pesanError != "" {
		a.tampilFormLamaran(w, r, item, true, pesanError)
		return
	}
	if _, err := a.Store.CreateFulltime(r.Context(), item); err != nil {
		a.serverError(w, r, err, "simpan lamaran")
		return
	}
	a.flash(r, "Lamaran ke "+item.Perusahaan+" berhasil disimpan.", "sukses")
	http.Redirect(w, r, "/lamaran", http.StatusSeeOther)
}

func (a *App) lamaranUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	item, pesanError := a.bacaFormLamaran(r)
	item.ID = id
	if pesanError != "" {
		a.tampilFormLamaran(w, r, item, false, pesanError)
		return
	}
	if err := a.Store.UpdateFulltime(r.Context(), item); err != nil {
		a.serverError(w, r, err, "update lamaran")
		return
	}
	a.flash(r, "Perubahan pada "+item.Perusahaan+" tersimpan.", "sukses")
	http.Redirect(w, r, "/lamaran", http.StatusSeeOther)
}

func (a *App) tampilFormLamaran(w http.ResponseWriter, r *http.Request, item model.Fulltime, baru bool, pesanError string) {
	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	judul := "Edit Lamaran"
	if baru {
		judul = "Tambah Lamaran"
	}
	w.WriteHeader(http.StatusUnprocessableEntity)
	a.render(w, r, "lamaran_form.html", judul, "lamaran", map[string]any{
		"Item": item, "Pilihan": pilihan, "Baru": baru, "Error": pesanError,
	})
}

func (a *App) lamaranHapus(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := a.Store.DeleteFulltime(r.Context(), id); err != nil {
		a.serverError(w, r, err, "hapus lamaran")
		return
	}
	a.flash(r, "Lamaran dihapus.", "sukses")
	http.Redirect(w, r, "/lamaran", http.StatusSeeOther)
}

func (a *App) lamaranPindahFreelance(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, err := a.Store.PindahKeFreelance(r.Context(), id); err != nil {
		a.serverError(w, r, err, "pindah lamaran ke freelance")
		return
	}
	a.flash(r, "Lamaran dipindahkan ke Freelance.", "sukses")
	http.Redirect(w, r, "/lamaran", http.StatusSeeOther)
}
