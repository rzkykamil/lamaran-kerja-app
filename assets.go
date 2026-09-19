package lamarankerja

import "embed"

// Assets berisi template HTML dan file statis, ikut dibundel ke dalam binary
// supaya aplikasi bisa dijalankan dari folder mana saja.
//
//go:embed templates static
var Assets embed.FS
