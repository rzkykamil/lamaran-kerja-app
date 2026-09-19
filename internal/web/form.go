package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Helper pembaca form. Semua mengembalikan zero value kalau input kosong,
// karena form HTML selalu mengirim string kosong untuk field yang tidak diisi.

func teks(r *http.Request, nama string) string {
	return strings.TrimSpace(r.PostFormValue(nama))
}

func bilangan(r *http.Request, nama string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(r.PostFormValue(nama)))
	return n
}

// uang membaca input rupiah yang boleh ditulis dengan titik/koma/spasi: "12.000.000".
func uang(r *http.Request, nama string) int64 {
	s := r.PostFormValue(nama)
	var b strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	n, _ := strconv.ParseInt(b.String(), 10, 64)
	return n
}

func tanggalWajib(r *http.Request, nama string) time.Time {
	if t, ok := parseTanggalForm(r.PostFormValue(nama)); ok {
		return t
	}
	return time.Now().Truncate(24 * time.Hour)
}

func tanggalOpsional(r *http.Request, nama string) *time.Time {
	if t, ok := parseTanggalForm(r.PostFormValue(nama)); ok {
		return &t
	}
	return nil
}

func parseTanggalForm(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func idDariURL(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}
