package web

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"

	"lamarankerja/internal/store"
)

type App struct {
	Store    *store.Store
	Sessions *scs.SessionManager
	Log      *log.Logger

	templates map[string]*template.Template
	static    http.Handler
}

func NewApp(st *store.Store, sessions *scs.SessionManager, assets fs.FS, logger *log.Logger) (*App, error) {
	templates, err := parsePages(assets)
	if err != nil {
		return nil, err
	}
	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}
	return &App{
		Store:     st,
		Sessions:  sessions,
		Log:       logger,
		templates: templates,
		static:    http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))),
	}, nil
}

// halaman adalah data yang selalu tersedia di setiap template.
type halaman struct {
	Judul      string
	Aktif      string
	Flash      string
	FlashJenis string
	Data       any
}

func (a *App) render(w http.ResponseWriter, r *http.Request, page, judul, aktif string, data any) {
	tmpl, ok := a.templates[page]
	if !ok {
		a.serverError(w, r, nil, "template "+page+" tidak ditemukan")
		return
	}

	isi := halaman{
		Judul:      judul,
		Aktif:      aktif,
		Flash:      a.Sessions.PopString(r.Context(), "flash"),
		FlashJenis: a.Sessions.PopString(r.Context(), "flash_jenis"),
		Data:       data,
	}
	if isi.Flash != "" && isi.FlashJenis == "" {
		isi.FlashJenis = "sukses"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", isi); err != nil {
		a.Log.Printf("render %s gagal: %v", page, err)
	}
}

// renderPartial dipakai untuk balasan HTMX yang hanya menukar sebagian halaman.
func (a *App) renderPartial(w http.ResponseWriter, page, blok string, data any) {
	tmpl, ok := a.templates[page]
	if !ok {
		http.Error(w, "template tidak ditemukan", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, blok, data); err != nil {
		a.Log.Printf("render partial %s/%s gagal: %v", page, blok, err)
	}
}

func (a *App) flash(r *http.Request, pesan, jenis string) {
	a.Sessions.Put(r.Context(), "flash", pesan)
	a.Sessions.Put(r.Context(), "flash_jenis", jenis)
}

func (a *App) serverError(w http.ResponseWriter, r *http.Request, err error, pesan string) {
	a.Log.Printf("error %s %s: %v (%s)", r.Method, r.URL.Path, err, pesan)
	http.Error(w, "Terjadi kesalahan di server. Coba lagi ya.", http.StatusInternalServerError)
}
