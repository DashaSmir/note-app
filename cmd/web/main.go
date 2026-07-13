package main
import (
    "fmt"
    "net/http"

    "github.com/go-chi/chi/v5"
)

func main() {
    r := chi.NewRouter()
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, world!")
    })

    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", r)
}