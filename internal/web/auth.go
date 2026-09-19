package web

import (
	"errors"
	"net/http"
	"strings"

	"lamarankerja/internal/store"
)

func (a *App) loginForm(w http.ResponseWriter, r *http.Request) {
	if a.Sessions.Exists(r.Context(), sesiUserID) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	a.renderPartial(w, "login.html", "login", map[string]any{
		"Flash": a.Sessions.PopString(r.Context(), "flash"),
	})
}

func (a *App) loginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(r.PostFormValue("username"))
	password := r.PostFormValue("password")

	user, err := a.Store.CekPassword(r.Context(), username, password)
	if err != nil {
		if errors.Is(err, store.ErrLoginSalah) {
			a.renderPartial(w, "login.html", "login", map[string]any{
				"Error":    "Username atau password salah.",
				"Username": username,
			})
			return
		}
		a.serverError(w, r, err, "cek password")
		return
	}

	// Ganti session ID setelah login untuk mencegah session fixation.
	if err := a.Sessions.RenewToken(r.Context()); err != nil {
		a.serverError(w, r, err, "renew token")
		return
	}
	a.Sessions.Put(r.Context(), sesiUserID, user.ID)
	a.Sessions.Put(r.Context(), "username", user.Username)
	a.flash(r, "Selamat datang kembali, "+user.Username+"!", "sukses")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if err := a.Sessions.Destroy(r.Context()); err != nil {
		a.serverError(w, r, err, "logout")
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
