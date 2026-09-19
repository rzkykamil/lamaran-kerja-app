# Tracker Lamaran Kerja & Freelance

Aplikasi web pengganti `tracker_lamaran_kamil.xlsx`. Dibuat dengan Go + PostgreSQL,
tanpa Node.js dan tanpa build step apa pun.

## Fitur

- **Login** satu akun (tidak ada registrasi publik)
- **Tracker Lamaran Full-Time** — CRUD, filter, pencarian, badge status berwarna
- **Tracker Project Freelance** — nilai tertimbang & sisa tagihan dihitung otomatis oleh database
- **Dashboard** — target vs aktual, funnel status, ringkasan uang, panel "Perlu Follow-up"
- **Export** ke Excel (.xlsx) dan CSV, mengikuti filter yang sedang aktif
- **Kelola Pilihan Dropdown** — pengganti sheet `Lists`
- **Import** data lama dari file Excel

## Setup Pertama Kali

### 1. Buat database

```bash
createdb -U postgres lamaran_kerja
```

Atau lewat psql:

```bash
psql -U postgres -c "CREATE DATABASE lamaran_kerja"
```

### 2. Isi konfigurasi

```bash
cp .env.example .env
```

Buka `.env`, ganti `PASSWORD_POSTGRES` dengan password user `postgres` di komputermu.

### 3. Buat akun login

```bash
go run ./cmd/seedadmin
```

Akan diminta username (default `admin`) dan password (minimal 6 karakter).
Tabel-tabel database dibuat otomatis pada langkah ini.

### 4. Import data dari Excel (opsional, sekali saja)

Tutup dulu file Excel-nya kalau sedang terbuka, lalu:

```bash
go run ./cmd/importxlsx "tracker_lamaran_kamil.xlsx"
```

Aman dijalankan berkali-kali — baris yang sudah ada akan dilewati.

### 5. Jalankan aplikasi

```bash
go run ./cmd/server
```

Buka http://localhost:8080

## Perintah Lain

| Perintah | Kegunaan |
|---|---|
| `go run ./cmd/seedadmin` | Ganti password kalau lupa (jalur pemulihan satu-satunya) |
| `go run ./cmd/seedadmin admin rahasia123` | Sama, tapi tanpa prompt |
| `go build -o tracker.exe ./cmd/server` | Bikin satu file .exe siap pakai |
| `curl localhost:8080/healthz` | Cek koneksi database |

Template HTML dan file statis ikut dibundel ke dalam binary, jadi `tracker.exe` bisa
dipindah ke folder mana pun asal `.env` ikut dibawa.

## Catatan Teknis

- **Migrasi** jalan otomatis saat startup dari `internal/database/migrations/*.sql`.
  Untuk menambah tabel, buat file `0003_*.sql` baru — jangan ubah file yang sudah pernah jalan.
- **`nilai_tertimbang` dan `sisa_tagihan`** adalah kolom `GENERATED` di PostgreSQL,
  jadi tidak pernah ditulis dari Go. Ubah `nilai_penawaran`/`probabilitas`/`sudah_dibayar`,
  angkanya ikut menyesuaikan.
- **Umur lamaran** dihitung saat query (`CURRENT_DATE - tgl_update_terakhir`),
  bukan disimpan — supaya tidak basi kalau seharian tidak ada perubahan data.
- **Pilihan dropdown** disimpan di tabel `list_items`. Menghapus sebuah pilihan tidak
  mengubah data lama yang terlanjur memakai nilai itu.
