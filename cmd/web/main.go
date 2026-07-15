package main
import (
    "database/sql"
    "fmt"
    "html/template"
    "log/slog"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "github.com/go-chi/chi/v5"
    _ "modernc.org/sqlite"
    "github.com/DashaSmir/note-app/internal/models"
)

type Config struct {
    Port string
    DSN  string
}

type application struct {
    config        Config
    logger        *slog.Logger
    notes         *models.NoteModel
    templateCache map[string]*template.Template
}

func main() {
    cfg := Config{
        Port: getEnv("PORT", "8080"),
        DSN:  getEnv("DSN", "./notes.db"),
    }

    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    db, err := sql.Open("sqlite", cfg.DSN)
    if err != nil {
        logger.Error("failed to open db", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS notes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        logger.Error("failed to create table", "error", err)
        os.Exit(1)
    }

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

//Загрузка всех шаблонов из папки internal/web/templates
func newTemplateCache() (map[string]*template.Template, error) {
    cache := map[string]*template.Template{}
    pages, err := filepath.Glob("./internal/web/templates/*.html")
    if err != nil {
        return nil, err
    }
    for _, page := range pages {
        name := filepath.Base(page)
        ts, err := template.ParseFiles("./internal/web/templates/base.html", page)
        if err != nil {
            return nil, err
        }
        cache[name] = ts
    }
    return cache, nil
}

//Рендеринг шаблона
func (app *application) render(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
    ts, ok := app.templateCache[name]
    if !ok {
        app.logger.Error("template not found", "name", name)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    err := ts.ExecuteTemplate(w, "base", data)
    if err != nil {
        app.logger.Error("template execution error", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
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

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    notes, err := app.notes.Latest()
    if err != nil {
        app.logger.Error("failed to get notes", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    data := struct{ Notes []*models.Note }{Notes: notes}
    app.render(w, r, "index.html", data)
}

func (app *application) viewNote(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }
    note, err := app.notes.Get(id)
    if err == models.ErrNoRecord {
        http.NotFound(w, r)
        return
    } else if err != nil {
        app.logger.Error("failed to get note", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    app.render(w, r, "view.html", note)
}

func (app *application) createNoteForm(w http.ResponseWriter, r *http.Request) {
    app.render(w, r, "create.html", nil)
}

func (app *application) createNotePost(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        app.logger.Error("failed to parse form", "error", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    title := r.PostForm.Get("title")
    content := r.PostForm.Get("content")

    if title == "" || content == "" {
        http.Error(w, "Title and content are required", http.StatusBadRequest)
        return
    }

    id, err := app.notes.Insert(title, content)
    if err != nil {
        app.logger.Error("failed to insert note", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/note/view/%d", id), http.StatusSeeOther)
}

func (app *application) editNoteForm(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }
    note, err := app.notes.Get(id)
    if err == models.ErrNoRecord {
        http.NotFound(w, r)
        return
    } else if err != nil {
        app.logger.Error("failed to get note", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    app.render(w, r, "edit.html", note)
}

func (app *application) editNotePost(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }

    err = r.ParseForm()
    if err != nil {
        app.logger.Error("failed to parse form", "error", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    title := r.PostForm.Get("title")
    content := r.PostForm.Get("content")

    if title == "" || content == "" {
        http.Error(w, "Title and content are required", http.StatusBadRequest)
        return
    }

    err = app.notes.Update(id, title, content)
    if err == models.ErrNoRecord {
        http.NotFound(w, r)
        return
    } else if err != nil {
        app.logger.Error("failed to update note", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/note/view/%d", id), http.StatusSeeOther)
}

func (app *application) deleteNote(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }

    err = app.notes.Delete(id)
    if err == models.ErrNoRecord {
        http.NotFound(w, r)
        return
    } else if err != nil {
        app.logger.Error("failed to delete note", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}