package validate

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrPasswordTooShort     = errors.New("password must be at least 8 characters")
	ErrEmailRequired        = errors.New("a valid email is required")
	ErrInvalidUUID          = errors.New("invalid id format")
	ErrBusinessNameRequired = errors.New("business_name is required")
)

func Password(p string) error {
	if len(p) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func Email(e string) error {
	if e == "" || !strings.Contains(e, "@") {
		return ErrEmailRequired
	}
	return nil
}

func UUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, ErrInvalidUUID
	}
	return id, nil
}
