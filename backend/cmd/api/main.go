package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"sencha/backend/internal/config"
	"sencha/backend/internal/handler"
	"sencha/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func loadConfig(path string) (*config.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("could not load config from %s/%s: %w", wd, path, err)
	}
	return cfg, nil
}

func initRepository(cfg *config.Config) (repository.Repository, error) {
	if cfg.DatabaseURL != "" {
		log.Printf("[main] connecting to postgres: %s", cfg.DatabaseURL)

		m, err := migrate.New(
			"file://internal/repository/migrations",
			cfg.DatabaseURL,
		)
		if err != nil {
			return nil, fmt.Errorf("creating migrator: %w", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, fmt.Errorf("running migrations: %w", err)
		}
		log.Println("[main] migrations applied")

		repo, err := repository.NewPostgres(context.Background(), cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("creating postgres repository: %w", err)
		}
		log.Println("[main] using postgres repository")
		return repo, nil
	}

	repo := repository.NewMemory()
	if err := repository.Seed(repo); err != nil {
		return nil, fmt.Errorf("seeding in-memory repository: %w", err)
	}
	log.Println("[main] using in-memory repository")
	return repo, nil
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("fatal: %v\n\nMake sure you run from the project root, e.g.:\n\tgo run ./cmd/api/\n", err)
	}

	repo, err := initRepository(cfg)
	if err != nil {
		log.Fatalf("fatal: %v\n", err)
	}
	cfg.Repository = repo

	handler.Initialize(cfg)

	r := gin.Default()
	handler.RegisterRoutes(r)
	fmt.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
