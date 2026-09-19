package model

import (
	"net/url"
	"strconv"
	"time"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type ListItem struct {
	ID        int64
	Category  string
	Value     string
	SortOrder int
}

type Fulltime struct {
	ID                int64
	TglApply          time.Time
	Perusahaan        string
	Posisi            string
	Lokasi            string
	TipeKerja         string
	Sumber            string
	LinkLowongan      string
	EkspektasiGaji    int64
	Prioritas         string
	Status            string
	TahapKe           int
	TglUpdateTerakhir time.Time
	NextAction        string
	TglFollowUp       *time.Time
	Kontak            string
	Catatan           string
	UmurHari          int
}

// FollowUpTelat true kalau tanggal follow-up sudah lewat dan lamarannya masih berjalan.
func (f Fulltime) FollowUpTelat() bool {
	if f.TglFollowUp == nil || f.Selesai() {
		return false
	}
	today := time.Now().Truncate(24 * time.Hour)
	return f.TglFollowUp.Before(today)
}

// Nganggur true kalau sudah lebih dari 14 hari tanpa update dan masih berjalan.
func (f Fulltime) Nganggur() bool {
	return !f.Selesai() && f.UmurHari > 14
}

func (f Fulltime) Selesai() bool {
	switch f.Status {
	case "Diterima", "Ditolak", "Withdraw", "Ghosting":
		return true
	}
	return false
}

type Freelance struct {
	ID               int64
	TglMasukLead     time.Time
	Klien            string
	NamaProject      string
	Stack            string
	Sumber           string
	ScopeSingkat     string
	NilaiPenawaran   int64
	Status           string
	Probabilitas     float64
	NilaiTertimbang  int64
	DeadlineProposal *time.Time
	TglMulai         *time.Time
	DeadlineProject  *time.Time
	StatusBayar      string
	SudahDibayar     int64
	SisaTagihan      int64
	NextAction       string
	Catatan          string
}

func (f Freelance) DeadlineTelat() bool {
	if f.DeadlineProject == nil || f.Status == "Selesai" || f.Status == "Batal" {
		return false
	}
	return f.DeadlineProject.Before(time.Now().Truncate(24 * time.Hour))
}

type FilterFulltime struct {
	Cari      string
	Status    string
	Prioritas string
	Sumber    string
	Urut      string
	Halaman   int
}

type FilterFreelance struct {
	Cari        string
	Status      string
	StatusBayar string
	Sumber      string
	Urut        string
	Halaman     int
}

// QueryHalaman membangun query string filter yang sedang aktif dengan nomor
// halaman diganti, supaya tombol paginasi tidak menghilangkan pencarian/filter.
func (f FilterFulltime) QueryHalaman(halaman int) string {
	v := url.Values{}
	setJikaAda(v, "cari", f.Cari)
	setJikaAda(v, "status", f.Status)
	setJikaAda(v, "prioritas", f.Prioritas)
	setJikaAda(v, "sumber", f.Sumber)
	setJikaAda(v, "urut", f.Urut)
	v.Set("halaman", strconv.Itoa(halaman))
	return "?" + v.Encode()
}

func (f FilterFreelance) QueryHalaman(halaman int) string {
	v := url.Values{}
	setJikaAda(v, "cari", f.Cari)
	setJikaAda(v, "status", f.Status)
	setJikaAda(v, "status_bayar", f.StatusBayar)
	setJikaAda(v, "sumber", f.Sumber)
	setJikaAda(v, "urut", f.Urut)
	v.Set("halaman", strconv.Itoa(halaman))
	return "?" + v.Encode()
}

func setJikaAda(v url.Values, kunci, nilai string) {
	if nilai != "" {
		v.Set(kunci, nilai)
	}
}

// Paginasi menyimpan info halaman untuk tabel yang panjang.
type Paginasi struct {
	Halaman      int
	PerHalaman   int
	Total        int
	TotalHalaman int
}

// HitungPaginasi membereskan nomor halaman yang aneh (0, negatif, atau
// kelewat besar) supaya selalu ada di rentang yang valid.
func HitungPaginasi(halaman, perHalaman, total int) Paginasi {
	if perHalaman < 1 {
		perHalaman = 1
	}
	totalHalaman := max((total+perHalaman-1)/perHalaman, 1)
	halaman = min(max(halaman, 1), totalHalaman)
	return Paginasi{Halaman: halaman, PerHalaman: perHalaman, Total: total, TotalHalaman: totalHalaman}
}

func (p Paginasi) AdaSebelumnya() bool { return p.Halaman > 1 }
func (p Paginasi) AdaBerikutnya() bool { return p.Halaman < p.TotalHalaman }
func (p Paginasi) Sebelumnya() int     { return p.Halaman - 1 }
func (p Paginasi) Berikutnya() int     { return p.Halaman + 1 }

// Dari dan Sampai dipakai untuk teks "menampilkan X-Y dari Z hasil".
func (p Paginasi) Dari() int {
	if p.Total == 0 {
		return 0
	}
	return (p.Halaman-1)*p.PerHalaman + 1
}

func (p Paginasi) Sampai() int {
	return min(p.Halaman*p.PerHalaman, p.Total)
}

type FunnelRow struct {
	Label  string
	Jumlah int
}

type TargetRow struct {
	Key      string
	Label    string
	Target   int
	Aktual   int
	Progress int
	Catatan  string
}

type Dashboard struct {
	Targets []TargetRow

	FunnelFulltime  []FunnelRow
	FunnelFreelance []FunnelRow

	TotalLamaran    int
	MasukInterview  int
	LolosInterview2 int
	OfferDiterima   int
	Nganggur        int
	ResponseRate    float64

	TotalPipeline   int64
	NilaiTertimbang int64
	NilaiClosing    int64
	UangMasuk       int64
	SisaTagihan     int64
	WinRate         float64
	ProposalKeluar  int
	FreelanceDeal   int

	PerluFollowUp []Fulltime
}
