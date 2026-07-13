package models

import (
    "database/sql"
    "time"
)

type Note struct {
    ID      int
    Title   string
    Content string
    Created time.Time
}

type NoteModel struct {
    DB *sql.DB
}

func (m *NoteModel) Insert(title, content string) (int, error) {
    stmt := `INSERT INTO notes (title, content, created) VALUES (?, ?, ?)`
    result, err := m.DB.Exec(stmt, title, content, time.Now())
    if err != nil {
        return 0, err
    }
    id, err := result.LastInsertId()
    return int(id), err
}

func (m *NoteModel) Get(id int) (*Note, error) {
    stmt := `SELECT id, title, content, created FROM notes WHERE id = ?`
    row := m.DB.QueryRow(stmt, id)
    n := &Note{}
    err := row.Scan(&n.ID, &n.Title, &n.Content, &n.Created)
    if err == sql.ErrNoRows {
        return nil, ErrNoRecord
    }
    return n, err
}
