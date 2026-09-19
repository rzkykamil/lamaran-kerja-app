package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"lamarankerja/internal/importer"
)

func (a *App) imporForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "impor.html", "Impor Data", "impor", map[string]any{})
}

func (a *App) imporSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		a.tampilImporError(w, r, "File terlalu besar atau form tidak valid.")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		a.tampilImporError(w, r, "Pilih file Excel (.xlsx) terlebih dahulu.")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		a.tampilImporError(w, r, "File harus berformat .xlsx.")
		return
	}

	tmp, err := os.CreateTemp("", "impor-*.xlsx")
	if err != nil {
		a.serverError(w, r, err, "buat file sementara impor")
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		a.serverError(w, r, err, "salin file impor")
		return
	}
	tmp.Close()

	hasil, err := importer.Import(r.Context(), a.Store, tmpPath)
	if err != nil {
		a.tampilImporError(w, r, "Gagal impor: "+err.Error())
		return
	}

	a.flash(r, fmt.Sprintf(
		"Impor selesai. Lamaran: %d masuk, %d dilewati. Freelance: %d masuk, %d dilewati.",
		hasil.LamaranMasuk, hasil.LamaranDilewati, hasil.ProjectMasuk, hasil.ProjectDilewati), "sukses")
	http.Redirect(w, r, "/impor", http.StatusSeeOther)
}

func (a *App) tampilImporError(w http.ResponseWriter, r *http.Request, pesan string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	a.render(w, r, "impor.html", "Impor Data", "impor", map[string]any{"Error": pesan})
}
