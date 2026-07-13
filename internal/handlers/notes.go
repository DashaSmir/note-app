package handlers
import (
    "html/template"
    "net/http"
    "strconv"
    "github.com/go-chi/chi/v5"
    "notes-app/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    notes, err := app.notes.Latest()
    if err != nil {
        app.logger.Error("failed to get notes", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    data := &templateData{Notes: notes}
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
    data := &templateData{Note: note}
    app.render(w, r, "view.html", data)
}