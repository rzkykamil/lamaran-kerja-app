package web

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

var headerLamaran = []string{"No", "Tgl Apply", "Perusahaan", "Posisi", "Lokasi", "Tipe Kerja",
	"Sumber", "Link Lowongan", "Ekspektasi Gaji (Rp)", "Prioritas", "Status", "Tahap Ke-",
	"Tgl Update Terakhir", "Umur (hari)", "Next Action", "Tgl Follow-up", "Kontak", "Catatan"}

var headerFreelance = []string{"No", "Tgl Masuk Lead", "Klien", "Nama Project", "Stack", "Sumber",
	"Scope Singkat", "Nilai Penawaran (Rp)", "Status", "Probabilitas", "Nilai Tertimbang (Rp)",
	"Deadline Proposal", "Tgl Mulai", "Deadline Project", "Status Bayar", "Sudah Dibayar (Rp)",
	"Sisa Tagihan (Rp)", "Next Action", "Catatan"}

func (a *App) lamaranExport(w http.ResponseWriter, r *http.Request) {
	items, err := a.Store.ListFulltimeSemua(r.Context(), filterLamaranDariURL(r))
	if err != nil {
		a.serverError(w, r, err, "export lamaran")
		return
	}

	baris := make([][]any, 0, len(items))
	for i, it := range items {
		baris = append(baris, []any{
			i + 1, tglExport(it.TglApply), it.Perusahaan, it.Posisi, it.Lokasi, it.TipeKerja,
			it.Sumber, it.LinkLowongan, it.EkspektasiGaji, it.Prioritas, it.Status, it.TahapKe,
			tglExport(it.TglUpdateTerakhir), it.UmurHari, it.NextAction, tglExportPtr(it.TglFollowUp),
			it.Kontak, it.Catatan,
		})
	}
	a.kirimExport(w, r, "lamaran-kerja", "Lamaran", headerLamaran, baris)
}

func (a *App) freelanceExport(w http.ResponseWriter, r *http.Request) {
	items, err := a.Store.ListFreelanceSemua(r.Context(), filterFreelanceDariURL(r))
	if err != nil {
		a.serverError(w, r, err, "export freelance")
		return
	}

	baris := make([][]any, 0, len(items))
	for i, it := range items {
		baris = append(baris, []any{
			i + 1, tglExport(it.TglMasukLead), it.Klien, it.NamaProject, it.Stack, it.Sumber,
			it.ScopeSingkat, it.NilaiPenawaran, it.Status, it.Probabilitas, it.NilaiTertimbang,
			tglExportPtr(it.DeadlineProposal), tglExportPtr(it.TglMulai), tglExportPtr(it.DeadlineProject),
			it.StatusBayar, it.SudahDibayar, it.SisaTagihan, it.NextAction, it.Catatan,
		})
	}
	a.kirimExport(w, r, "project-freelance", "Freelance", headerFreelance, baris)
}

func (a *App) kirimExport(w http.ResponseWriter, r *http.Request, namaFile, sheet string, header []string, baris [][]any) {
	stamp := time.Now().Format("2006-01-02")

	if r.URL.Query().Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename=%s-%s.csv", namaFile, stamp))

		// BOM supaya Excel membaca huruf beraksen dengan benar.
		w.Write([]byte{0xEF, 0xBB, 0xBF})
		cw := csv.NewWriter(w)
		cw.Write(header)
		for _, row := range baris {
			teks := make([]string, len(row))
			for i, v := range row {
				teks[i] = keString(v)
			}
			cw.Write(teks)
		}
		cw.Flush()
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Family: "Arial", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"2563EB"}},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})

	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	if first, err := excelize.CoordinatesToCellName(1, 1); err == nil {
		if last, err := excelize.CoordinatesToCellName(len(header), 1); err == nil {
			f.SetCellStyle(sheet, first, last, headerStyle)
			f.SetRowHeight(sheet, 1, 28)
			f.AutoFilter(sheet, first+":"+last, nil)
		}
	}

	for r0, row := range baris {
		for c0, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c0+1, r0+2)
			f.SetCellValue(sheet, cell, v)
		}
	}

	for i := range header {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 20)
	}
	f.SetPanes(sheet, &excelize.Panes{
		Freeze: true, Split: false, XSplit: 0, YSplit: 1,
		TopLeftCell: "A2", ActivePane: "bottomLeft",
	})

	w.Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s-%s.xlsx", namaFile, stamp))
	if err := f.Write(w); err != nil {
		a.Log.Printf("gagal menulis file export: %v", err)
	}
}

func tglExport(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func tglExportPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return tglExport(*t)
}

func keString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}
