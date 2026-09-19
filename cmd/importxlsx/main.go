// Command importxlsx memindahkan isi tracker_lamaran_kamil.xlsx ke database.
// Aman dijalankan berkali-kali: baris yang sudah ada akan dilewati.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"lamarankerja/internal/config"
	"lamarankerja/internal/database"
	"lamarankerja/internal/importer"
	"lamarankerja/internal/store"
)

func main() {
	path := "tracker_lamaran_kamil.xlsx"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Membaca %s ...\n", path)
	hasil, err := importer.Import(ctx, store.New(pool), path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nSelesai.\n  Lamaran  : %d masuk, %d dilewati\n  Freelance: %d masuk, %d dilewati\n",
		hasil.LamaranMasuk, hasil.LamaranDilewati, hasil.ProjectMasuk, hasil.ProjectDilewati)
}
