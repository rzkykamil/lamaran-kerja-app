// Command seedadmin membuat atau mengganti password akun admin.
// Ini juga jalur pemulihan kalau password lupa: jalankan ulang dengan password baru.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"lamarankerja/internal/config"
	"lamarankerja/internal/database"
	"lamarankerja/internal/store"
)

func main() {
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

	username, password := bacaKredensial()
	if err := store.New(pool).SimpanAdmin(ctx, username, password); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nAkun \"%s\" siap dipakai. Login di http://localhost:%s/login\n", username, cfg.Port)
}

// Kredensial bisa lewat argumen (dipakai di skrip) atau diketik interaktif.
func bacaKredensial() (string, string) {
	if len(os.Args) >= 3 {
		return os.Args[1], os.Args[2]
	}

	in := bufio.NewReader(os.Stdin)
	fmt.Print("Username [admin]: ")
	username, _ := in.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	fmt.Print("Password: ")
	password, _ := in.ReadString('\n')
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		log.Fatal("password minimal 6 karakter")
	}
	return username, password
}
