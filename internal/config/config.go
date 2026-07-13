package config
type Config struct {
    Port     string
    DSN      string
}

// cmd/web/main.go
package main

import (
    "database/sql"
    "log/slog"
    "net/http"
    "os"
    "github.com/go-chi/chi/v5"
    "github.com/mattn/go-sqlite3"
    "notes-app/internal/config"
    "notes-app/internal/handlers"
    "notes-app/internal/models"
)

type application struct {
    config    config.Config
    logger    *slog.Logger
    notes     *models.NoteModel
}

func main() {
    cfg := config.Config{
        Port: os.Getenv("PORT"),
        DSN:  os.Getenv("DSN"),
    }
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    db, err := sql.Open("sqlite3", cfg.DSN)
    if err != nil {
        logger.Error("failed to open db", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    app := &application{
        config: cfg,
        logger: logger,
        notes:  &models.NoteModel{DB: db},
    }

    r := chi.NewRouter()
    app.routes(r) //этот метод определим в routes.go

    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: r,
    }

    logger.Info("starting server", "port", cfg.Port)
    if err := srv.ListenAndServe(); err != nil {
        logger.Error("server error", "error", err)
        os.Exit(1)
    }
}