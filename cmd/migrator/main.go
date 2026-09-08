package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/zlatovlaska707/avito-kitchen/internal/config"
)

func main() {
	var configPath string
	var withSeed bool
	flag.StringVar(&configPath, "config", "config/config.yaml", "path to config yaml")
	flag.BoolVar(&withSeed, "seed", false, "run seed.sql after up")
	flag.Parse()

	cfg := config.MustLoadMigrator(configPath)
	args := flag.Args()
	// Support both "./migrator --seed up" and "./migrator up --seed".
	for i, arg := range args {
		if arg == "--seed" {
			withSeed = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: migrator [--seed] [up|down|down-to|status|seed]")
		os.Exit(1)
	}

	if err := run(args, cfg, withSeed); err != nil {
		fmt.Fprintf(os.Stderr, "migrator: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, cfg config.MigratorConfig, withSeed bool) error {
	db, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	dir := cfg.Migrate.Dir
	command := args[0]
	switch command {
	case "seed":
		return runSeed(db, dir)
	case "up", "down", "down-to", "status", "redo", "reset", "version":
		if err := goose.RunContext(context.Background(), command, db, dir, args[1:]...); err != nil {
			return err
		}
		if withSeed && command == "up" {
			return runSeed(db, dir)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func runSeed(db *sql.DB, migrationsDir string) error {
	seedPath := filepath.Join(migrationsDir, "seed", "seed.sql")
	content, err := os.ReadFile(seedPath) //nolint:gosec // seed path is derived from configured migrations dir
	if err != nil {
		return fmt.Errorf("read seed: %w", err)
	}
	if _, err := db.ExecContext(context.Background(), string(content)); err != nil {
		return fmt.Errorf("exec seed: %w", err)
	}
	fmt.Println("seed applied")
	return nil
}

func init() {
	stdlib.GetDefaultDriver()
}
