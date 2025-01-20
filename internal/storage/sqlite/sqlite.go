package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"URL-shortener/internal/domain"
	"URL-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

const urlCreatedType = "URLCreated"

func New(dbPath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%s: open db: %w", op, err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS url (
			id INTEGER PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_alias ON url(alias)`,
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY,
			event_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'done')),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return nil, fmt.Errorf("%s: exec query: %w", op, err)
		}
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (id int64, err error) {
	const op = "storage.sqlite.SaveURL"

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s transaction %w", op, err)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("%s commit %w", op, commitErr)
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO url(url, alias) values (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s prepare statement %w", op, err)
	}
	defer stmt.Close()

	resp, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: unique constraint failed: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s execute statement %w", op, err)
	}

	id, err = resp.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id %w", op, err)
	}

	eventPayload := fmt.Sprintf(
		"id: %d, url: %s, alias: %s",
		id,
		urlToSave,
		alias,
	)

	if err = s.saveEvent(tx, urlCreatedType, eventPayload); err != nil {
		return 0, fmt.Errorf("%s failed to save event %w", op, err)
	}

	return id, nil
}

func (s *Storage) saveEvent(tx *sql.Tx, eventType string, payload string) error {
	const op = "storage.sqlite.saveEvent"

	stmt, err := tx.Prepare("INSERT INTO events(event_type, payload) values(?,?) ")
	if err != nil {
		return fmt.Errorf("%s prepare statement %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(eventType, payload)
	if err != nil {
		return fmt.Errorf("%s failed to execute statement %w", op, err)
	}

	return nil
}

type event struct {
	Id      int    `db:"id"`
	Type    string `db:"event_type"`
	Payload string `db:"payload"`
}

func (s *Storage) GetNewEvent() (domain.Event, error) {
	const op = "storage.sqlite.GetNewEvent"

	row := s.db.QueryRow("SELECT id, event_type, payload FROM events WHERE status = 'new' LIMIT 1")

	var evt event

	if err := row.Scan(&evt.Id, &evt.Type, &evt.Payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, nil
		}

		return domain.Event{}, fmt.Errorf("%s failed to scan row %w", op, err)
	}

	return domain.Event{
		ID:      evt.Id,
		Type:    evt.Type,
		Payload: evt.Payload,
	}, nil
}

func (s *Storage) SetEventDone(id int) error {
	const op = "storage.sqlite.SetEventDone"

	stmt, err := s.db.Prepare("UPDATE events SET status = 'done' WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s failed to prepare statement %w", op, err)
	}

	if _, err = stmt.Exec(id); err != nil {
		return fmt.Errorf("%s failed to execute statement %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(aliasToFind string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s prepare statement %w", op, err)
	}
	defer stmt.Close()

	var resUrl string
	err = stmt.QueryRow(aliasToFind).Scan(&resUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s execute statement %w", op, err)
	}

	return resUrl, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	res, err := s.db.Exec("DELETE FROM url WHERE alias = ?", alias)
	if err != nil {
		return fmt.Errorf("%s: exec delete: %w", op, err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: rowsAffected: %w", op, err)
	}

	return nil
}
