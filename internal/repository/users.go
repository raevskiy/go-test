package repository

import (
	"context"
	"cruder/internal/model"
	"database/sql"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	DeleteByUuid(uuid uuid.UUID) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.QueryContext(
		context.Background(),
		`SELECT id, uuid, username, email, full_name FROM users ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usersCount int
	err = r.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users").Scan(&usersCount)
	if err != nil {
		return nil, err
	}

	allUsers := make([]model.User, 0, usersCount)
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.FullName); err != nil {
			return nil, err
		}
		allUsers = append(allUsers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return allUsers, nil
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User

	if err := r.db.QueryRowContext(
		context.Background(),
		`SELECT id, uuid, username, email, full_name FROM users WHERE username = $1`,
		username).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.FullName); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(id int64) (*model.User, error) {
	var user model.User

	if err := r.db.QueryRowContext(
		context.Background(),
		`SELECT id, uuid, username, email, full_name FROM users WHERE id = $1`,
		id).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.FullName); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) DeleteByUuid(uuid uuid.UUID) error {
	res, err := r.db.ExecContext(context.Background(), `DELETE FROM users WHERE uuid = $1`, uuid)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
