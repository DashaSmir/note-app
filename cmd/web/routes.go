package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes(r *chi.Mux) {
    r.Use(middleware.Logger)
    r.Use(app.recoverPanic)

    fileServer := http.FileServer(http.Dir("./internal/web"))
    r.Handle("/css/*", http.StripPrefix("/css", fileServer))

    r.Get("/", app.home)
    r.Get("/note/view/{id}", app.viewNote)
    r.Get("/note/create", app.createNoteForm)
    r.Post("/note/create", app.createNotePost)
    r.Get("/note/edit/{id}", app.editNoteForm)
    r.Post("/note/edit/{id}", app.editNotePost)
    r.Post("/note/delete/{id}", app.deleteNote)
}