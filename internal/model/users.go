package model

import (
	"database/sql"
	"github.com/google/uuid"
)

type User struct {
	ID       int
	UUID     uuid.UUID
	Username string
	Email    string
	FullName sql.NullString
}
