package storage

import "errors"

var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrNoRecordAffected    = errors.New("no record affected by query")
	ErrRecordAlreadyExists = errors.New("record already exists")
)
