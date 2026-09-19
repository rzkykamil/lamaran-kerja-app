package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var namaBulan = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun",
	"Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// Rupiah memformat 12000000 menjadi "Rp 12.000.000".
func Rupiah(n int64) string {
	if n == 0 {
		return "Rp 0"
	}
	tanda := ""
	if n < 0 {
		tanda = "-"
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return tanda + "Rp " + b.String()
}

// RupiahSingkat memformat nominal besar jadi "12,0 jt" untuk kartu dashboard.
func RupiahSingkat(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return strings.Replace(fmt.Sprintf("Rp %.1f M", float64(n)/1e9), ".", ",", 1)
	case n >= 1_000_000:
		return strings.Replace(fmt.Sprintf("Rp %.1f jt", float64(n)/1e6), ".", ",", 1)
	case n >= 1_000:
		return strings.Replace(fmt.Sprintf("Rp %.0f rb", float64(n)/1e3), ".", ",", 1)
	default:
		return Rupiah(n)
	}
}

func tanggal(t any) string {
	tt, ok := keTime(t)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%d %s %d", tt.Day(), namaBulan[tt.Month()-1], tt.Year())
}

// tanggalInput menghasilkan format yang dimengerti <input type="date">.
func tanggalInput(t any) string {
	tt, ok := keTime(t)
	if !ok {
		return ""
	}
	return tt.Format("2006-01-02")
}

func keTime(v any) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, !t.IsZero()
	case *time.Time:
		if t == nil || t.IsZero() {
			return time.Time{}, false
		}
		return *t, true
	}
	return time.Time{}, false
}

func persen(f float64) string {
	return strings.Replace(fmt.Sprintf("%.0f%%", math.Round(f*100)), ".", ",", 1)
}

// slugStatus mengubah "Interview User" jadi "interview-user" untuk dipakai sebagai kelas CSS.
func slugStatus(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '/', r == '-':
			b.WriteRune('-')
		}
	}
	return strings.Trim(strings.ReplaceAll(b.String(), "--", "-"), "-")
}

var fungsiTemplate = template.FuncMap{
	"rupiah":        Rupiah,
	"rupiahSingkat": RupiahSingkat,
	"tanggal":       tanggal,
	"tanggalInput":  tanggalInput,
	"persen":        persen,
	"slug":          slugStatus,
	"angka":         func(n int) string { return strconv.Itoa(n) },
	"probPersen":    func(f float64) string { return strconv.Itoa(int(math.Round(f * 100))) },
	"add":           func(a, b int) int { return a + b },
	"dict": func(values ...any) (map[string]any, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict: jumlah argumen harus genap")
		}
		m := make(map[string]any, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict: key harus string")
			}
			m[key] = values[i+1]
		}
		return m, nil
	},
	"barWidth": func(n, max int) string {
		if max <= 0 {
			return "0%"
		}
		w := float64(n) / float64(max) * 100
		if n > 0 && w < 3 {
			w = 3
		}
		return fmt.Sprintf("%.1f%%", w)
	},
}

// parsePages membangun satu set template per halaman: layout + semua partial + halaman itu.
func parsePages(files fs.FS) (map[string]*template.Template, error) {
	pages, err := fs.Glob(files, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}
	out := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		name := filepath.Base(page)
		tmpl, err := template.New(name).Funcs(fungsiTemplate).ParseFS(files,
			"templates/layout.html", "templates/partials/*.html", page)
		if err != nil {
			return nil, fmt.Errorf("gagal parse template %s: %w", name, err)
		}
		out[name] = tmpl
	}
	return out, nil
}
