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
// Latest возвращает последние 10 заметок
func (m *NoteModel) Latest() ([]*Note, error) {
    stmt := `SELECT id, title, content, created FROM notes ORDER BY id DESC LIMIT 10`
    rows, err := m.DB.Query(stmt)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    notes := []*Note{}
    for rows.Next() {
        n := &Note{}
        err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Created)
        if err != nil {
            return nil, err
        }
        notes = append(notes, n)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }
    return notes, nil
}

// Update обновляет заголовок и содержание заметки
func (m *NoteModel) Update(id int, title, content string) error {
    stmt := `UPDATE notes SET title = ?, content = ? WHERE id = ?`
    _, err := m.DB.Exec(stmt, title, content, id)
    return err
}

// Delete удаляет заметку по ID
func (m *NoteModel) Delete(id int) error {
    stmt := `DELETE FROM notes WHERE id = ?`
    _, err := m.DB.Exec(stmt, id)
    return err
}