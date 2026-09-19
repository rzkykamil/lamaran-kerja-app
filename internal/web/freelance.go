package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"lamarankerja/internal/model"
)

func filterFreelanceDariURL(r *http.Request) model.FilterFreelance {
	q := r.URL.Query()
	halaman, _ := strconv.Atoi(q.Get("halaman"))
	return model.FilterFreelance{
		Cari:        q.Get("cari"),
		Status:      q.Get("status"),
		StatusBayar: q.Get("status_bayar"),
		Sumber:      q.Get("sumber"),
		Urut:        q.Get("urut"),
		Halaman:     halaman,
	}
}

func (a *App) freelanceList(w http.ResponseWriter, r *http.Request) {
	filter := filterFreelanceDariURL(r)
	items, paginasi, err := a.Store.ListFreelance(r.Context(), filter)
	if err != nil {
		a.serverError(w, r, err, "list freelance")
		return
	}

	// Total di kaki tabel dihitung dari SEMUA baris yang cocok filter, bukan
	// cuma yang tampil di halaman ini.
	totalNilai, totalWeighted, totalSisa, err := a.Store.AgregatFreelance(r.Context(), filter)
	if err != nil {
		a.serverError(w, r, err, "agregat freelance")
		return
	}

	data := map[string]any{
		"Items": items, "Filter": filter, "Paginasi": paginasi,
		"TotalNilai": totalNilai, "TotalWeighted": totalWeighted, "TotalSisa": totalSisa,
	}

	if r.Header.Get("HX-Request") == "true" {
		a.renderPartial(w, "freelance_list.html", "tabel-freelance", data)
		return
	}

	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	data["Pilihan"] = pilihan
	a.render(w, r, "freelance_list.html", "Project Freelance", "freelance", data)
}

func (a *App) freelanceBaru(w http.ResponseWriter, r *http.Request) {
	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	item := model.Freelance{
		TglMasukLead: time.Now().Truncate(24 * time.Hour),
		Status:       "Lead",
		StatusBayar:  "Belum Ditagih",
		Probabilitas: 0.5,
	}
	a.render(w, r, "freelance_form.html", "Tambah Project", "freelance", map[string]any{
		"Item": item, "Pilihan": pilihan, "Baru": true,
	})
}

func (a *App) freelanceEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	item, err := a.Store.GetFreelance(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	a.render(w, r, "freelance_form.html", "Edit Project", "freelance", map[string]any{
		"Item": item, "Pilihan": pilihan, "Baru": false,
	})
}

func (a *App) bacaFormFreelance(r *http.Request) (model.Freelance, string) {
	// Probabilitas diinput sebagai persen (0-100), disimpan sebagai pecahan 0-1.
	prob, _ := strconv.ParseFloat(strings.TrimSpace(r.PostFormValue("probabilitas")), 64)
	prob = prob / 100
	if prob < 0 {
		prob = 0
	}
	if prob > 1 {
		prob = 1
	}

	f := model.Freelance{
		TglMasukLead:     tanggalWajib(r, "tgl_masuk_lead"),
		Klien:            teks(r, "klien"),
		NamaProject:      teks(r, "nama_project"),
		Stack:            teks(r, "stack"),
		Sumber:           teks(r, "sumber"),
		ScopeSingkat:     teks(r, "scope_singkat"),
		NilaiPenawaran:   uang(r, "nilai_penawaran"),
		Status:           teks(r, "status"),
		Probabilitas:     prob,
		DeadlineProposal: tanggalOpsional(r, "deadline_proposal"),
		TglMulai:         tanggalOpsional(r, "tgl_mulai"),
		DeadlineProject:  tanggalOpsional(r, "deadline_project"),
		StatusBayar:      teks(r, "status_bayar"),
		SudahDibayar:     uang(r, "sudah_dibayar"),
		NextAction:       teks(r, "next_action"),
		Catatan:          teks(r, "catatan"),
	}
	// Dipakai kalau form harus ditampilkan ulang karena error; angka yang benar
	// tetap datang dari kolom GENERATED di PostgreSQL setelah tersimpan.
	f.NilaiTertimbang = int64(float64(f.NilaiPenawaran) * f.Probabilitas)
	f.SisaTagihan = f.NilaiPenawaran - f.SudahDibayar

	switch {
	case f.Klien == "":
		return f, "Nama klien wajib diisi."
	case f.NamaProject == "":
		return f, "Nama project wajib diisi."
	case f.SudahDibayar > f.NilaiPenawaran:
		return f, "Jumlah yang sudah dibayar tidak boleh melebihi nilai penawaran."
	}
	return f, ""
}

func (a *App) freelanceCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	item, pesanError := a.bacaFormFreelance(r)
	if pesanError != "" {
		a.tampilFormFreelance(w, r, item, true, pesanError)
		return
	}
	if _, err := a.Store.CreateFreelance(r.Context(), item); err != nil {
		a.serverError(w, r, err, "simpan freelance")
		return
	}
	a.flash(r, "Project "+item.NamaProject+" berhasil disimpan.", "sukses")
	http.Redirect(w, r, "/freelance", http.StatusSeeOther)
}

func (a *App) freelanceUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	item, pesanError := a.bacaFormFreelance(r)
	item.ID = id
	if pesanError != "" {
		a.tampilFormFreelance(w, r, item, false, pesanError)
		return
	}
	if err := a.Store.UpdateFreelance(r.Context(), item); err != nil {
		a.serverError(w, r, err, "update freelance")
		return
	}
	a.flash(r, "Perubahan pada "+item.NamaProject+" tersimpan.", "sukses")
	http.Redirect(w, r, "/freelance", http.StatusSeeOther)
}

func (a *App) tampilFormFreelance(w http.ResponseWriter, r *http.Request, item model.Freelance, baru bool, pesanError string) {
	pilihan, err := a.Store.SemuaPilihan(r.Context())
	if err != nil {
		a.serverError(w, r, err, "ambil pilihan")
		return
	}
	judul := "Edit Project"
	if baru {
		judul = "Tambah Project"
	}
	w.WriteHeader(http.StatusUnprocessableEntity)
	a.render(w, r, "freelance_form.html", judul, "freelance", map[string]any{
		"Item": item, "Pilihan": pilihan, "Baru": baru, "Error": pesanError,
	})
}

func (a *App) freelanceHapus(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := a.Store.DeleteFreelance(r.Context(), id); err != nil {
		a.serverError(w, r, err, "hapus freelance")
		return
	}
	a.flash(r, "Project dihapus.", "sukses")
	http.Redirect(w, r, "/freelance", http.StatusSeeOther)
}

func (a *App) freelancePindahLamaran(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariURL(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, err := a.Store.PindahKeLamaran(r.Context(), id); err != nil {
		a.serverError(w, r, err, "pindah freelance ke lamaran")
		return
	}
	a.flash(r, "Project dipindahkan ke Lamaran Kerja.", "sukses")
	http.Redirect(w, r, "/freelance", http.StatusSeeOther)
}
