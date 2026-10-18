package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Akarshit7/competition-platform-go/internal/config"
	"github.com/Akarshit7/competition-platform-go/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.MustLoad()
	pool, err := database.NewPool(cfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	email := "admin@competitions-social-media-platform.com"
	password := "Admin@123"
	name := "Super Admin"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash: %v", err)
	}

	_, err = pool.Exec(context.Background(),
		`INSERT INTO admins (email, password_hash, name) VALUES ($1,$2,$3) ON CONFLICT (email) DO NOTHING`,
		email, string(hash), name)
	if err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	fmt.Println("seeded admin:", email, "password:", password)
}
