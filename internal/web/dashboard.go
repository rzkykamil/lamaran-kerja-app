package web

import (
	"net/http"
	"strconv"
)

func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	d, err := a.Store.Dashboard(r.Context())
	if err != nil {
		a.serverError(w, r, err, "hitung dashboard")
		return
	}

	maxFunnel := 1
	for _, row := range d.FunnelFulltime {
		if row.Jumlah > maxFunnel {
			maxFunnel = row.Jumlah
		}
	}
	maxPipeline := 1
	for _, row := range d.FunnelFreelance {
		if row.Jumlah > maxPipeline {
			maxPipeline = row.Jumlah
		}
	}

	a.render(w, r, "dashboard.html", "Dashboard", "dashboard", map[string]any{
		"D":           d,
		"MaxFunnel":   maxFunnel,
		"MaxPipeline": maxPipeline,
		"Username":    a.Sessions.GetString(r.Context(), "username"),
	})
}

var kunciTarget = []string{"perusahaan_dilamar", "masuk_interview", "lolos_interview2",
	"offer_diterima", "freelance_closing"}

func (a *App) simpanTarget(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	for _, key := range kunciTarget {
		raw := r.PostFormValue(key)
		if raw == "" {
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			continue
		}
		if err := a.Store.SimpanTarget(r.Context(), key, n); err != nil {
			a.serverError(w, r, err, "simpan target")
			return
		}
	}
	a.flash(r, "Target berhasil diperbarui.", "sukses")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
