package main

import (
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-sqlite3"
	"github.com/DashaSmir/note-app/internal/models"
)

type application struct {
	config        Config
	logger        *slog.Logger
	notes         *models.NoteModel
	templateCache map[string]*template.Template
}

type Config struct {
	Port string
	DSN  string
}

func main() {
	cfg := Config{
		Port: getEnv("PORT", "8080"),
		DSN:  getEnv("DSN", "./notes.db"),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := sql.Open("sqlite3", cfg.DSN)
	if err != nil {
		logger.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("failed to load templates", "error", err)
		os.Exit(1)
	}

	app := &application{
		config:        cfg,
		logger:        logger,
		notes:         &models.NoteModel{DB: db},
		templateCache: templateCache,
	}

	r := chi.NewRouter()
	app.routes(r)

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)
	return cache, nil
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Пока просто пишем "Home"
	w.Write([]byte("Home page – список заметок скоро будет"))
}

func (app *application) viewNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Write([]byte("View note with id: " + id))
}

func (app *application) createNoteForm(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Форма создания заметки"))
}

func (app *application) createNotePost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Создание заметки (POST)"))
}

func (app *application) editNoteForm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Write([]byte("Форма редактирования заметки " + id))
}

func (app *application) editNotePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Write([]byte("Обновление заметки " + id + " (POST)"))
}

func (app *application) deleteNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Write([]byte("Удаление заметки " + id))
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}