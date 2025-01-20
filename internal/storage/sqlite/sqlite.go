package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"URL-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(dbPath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%s: open db: %w", op, err)
	}

	stmt, err := db.Prepare(`
    CREATE TABLE IF NOT EXISTS url (
        id INTEGER PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
    `)

	if err != nil {
		return nil, fmt.Errorf("%s prepare statement %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s execute statement %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) values (?, ?)")
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

	id, err := resp.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id %w", op, err)
	}

	return id, nil
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
