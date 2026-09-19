package web

import "net/http"

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", a.static)
	mux.HandleFunc("GET /healthz", a.healthz)
	mux.HandleFunc("GET /login", a.loginForm)
	mux.HandleFunc("POST /login", a.loginSubmit)

	// Semua route di bawah ini wajib login.
	privat := http.NewServeMux()
	privat.HandleFunc("POST /logout", a.logout)

	privat.HandleFunc("GET /{$}", a.dashboard)
	privat.HandleFunc("POST /dashboard/target", a.simpanTarget)

	privat.HandleFunc("GET /lamaran", a.lamaranList)
	privat.HandleFunc("GET /lamaran/baru", a.lamaranBaru)
	privat.HandleFunc("POST /lamaran", a.lamaranCreate)
	privat.HandleFunc("GET /lamaran/{id}/edit", a.lamaranEdit)
	privat.HandleFunc("POST /lamaran/{id}", a.lamaranUpdate)
	privat.HandleFunc("POST /lamaran/{id}/hapus", a.lamaranHapus)
	privat.HandleFunc("POST /lamaran/{id}/pindah-freelance", a.lamaranPindahFreelance)
	privat.HandleFunc("GET /lamaran/export", a.lamaranExport)

	privat.HandleFunc("GET /freelance", a.freelanceList)
	privat.HandleFunc("GET /freelance/baru", a.freelanceBaru)
	privat.HandleFunc("POST /freelance", a.freelanceCreate)
	privat.HandleFunc("GET /freelance/{id}/edit", a.freelanceEdit)
	privat.HandleFunc("POST /freelance/{id}", a.freelanceUpdate)
	privat.HandleFunc("POST /freelance/{id}/hapus", a.freelanceHapus)
	privat.HandleFunc("POST /freelance/{id}/pindah-lamaran", a.freelancePindahLamaran)
	privat.HandleFunc("GET /freelance/export", a.freelanceExport)

	privat.HandleFunc("GET /impor", a.imporForm)
	privat.HandleFunc("POST /impor", a.imporSubmit)

	privat.HandleFunc("GET /pilihan", a.pilihanList)
	privat.HandleFunc("POST /pilihan", a.pilihanTambah)
	privat.HandleFunc("POST /pilihan/{id}/hapus", a.pilihanHapus)

	mux.Handle("/", a.requireLogin(privat))

	return a.recoverPanic(a.logRequest(a.Sessions.LoadAndSave(mux)))
}

func (a *App) healthz(w http.ResponseWriter, r *http.Request) {
	if _, err := a.Store.JumlahUser(r.Context()); err != nil {
		http.Error(w, "database tidak terhubung: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Write([]byte("ok"))
}
