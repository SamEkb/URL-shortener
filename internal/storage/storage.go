package storage

import "errors"

var (
	ErrUrlNotFound = errors.New("url not found")
	ErrUrlExists   = errors.New("url exists")
)

type Storage interface {
	SaveURL(urlToSave string, alias string) (int64, error)
	GetURL(aliasToFind string) (string, error)
	DeleteURL(alias string) (bool, error)
}
