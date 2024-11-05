package pgerr

import (
	"database/sql"
	"errors"

	"github.com/fydmer/fileserver/internal/domain/repository"
	"github.com/lib/pq"
)

func Parse(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrResourceNotFound
	}

	pgErr := &pq.Error{}
	if errors.As(err, &pgErr) {
		switch pgErr.Code.Name() {
		case "not_null_violation":
			fallthrough
		case "foreign_key_violation":
			fallthrough
		case "check_violation":
			return repository.ErrBadRequest
		case "unique_violation":
			fallthrough
		case "exclusion_violation":
			return repository.ErrResourceAlreadyExists
		default:
			return repository.ErrUnknown
		}
	}

	return err
}
