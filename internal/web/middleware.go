package web

import (
	"net/http"
	"time"
)

const sesiUserID = "user_id"

// requireLogin menolak request yang belum punya sesi login.
func (a *App) requireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Sessions.Exists(r.Context(), sesiUserID) {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (a *App) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulai := time.Now()
		next.ServeHTTP(w, r)
		a.Log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(mulai).Round(time.Millisecond))
	})
}

func (a *App) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Connection", "close")
				a.Log.Printf("PANIC %s %s: %v", r.Method, r.URL.Path, rec)
				http.Error(w, "Terjadi kesalahan di server.", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
