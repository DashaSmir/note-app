func (app *application) routes(r *chi.Mux) {
    fileServer := http.FileServer(http.Dir("./internal/web/static"))
    r.Handle("/static/*", http.StripPrefix("/static", fileServer))
    r.Use(middleware.Logger)
    r.Use(app.recoverPanic)
    r.Get("/", app.home)
    r.Get("/note/view/{id}", app.viewNote)
    r.Get("/note/create", app.createNoteForm)
    r.Post("/note/create", app.createNotePost)
    r.Get("/note/edit/{id}", app.editNoteForm)
    r.Post("/note/edit/{id}", app.editNotePost)
    r.Post("/note/delete/{id}", app.deleteNote)
}