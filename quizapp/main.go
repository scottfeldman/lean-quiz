package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
)

func main() {
	addr := "127.0.0.1:8787"
	if v := os.Getenv("LEAN_QUIZ_ADDR"); v != "" {
		addr = v
	}
	dir, err := appSupportDir()
	if err != nil {
		log.Fatal(err)
	}
	st, err := loadStore(filepath.Join(dir, "state.json"))
	if err != nil {
		log.Fatal(err)
	}
	app := newQuizApp(st, nil)
	log.Printf("lean quiz listening on http://%s", addr)
	if err := app.fb.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		log.Fatal(err)
	}
}

func appSupportDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "Application Support", "LeanQuiz")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}
