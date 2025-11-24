package repository

import (
	"context"
	"cruder/internal/controller/dto"
	"cruder/internal/model"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"strings"
)

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	DeleteByUuid(uuid uuid.UUID) error
	PartiallyUpdateByUUID(uuid uuid.UUID, patch dto.UserPatch) error
	Create(user dto.UserCreate) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

var BusinessErrNoUsers = errors.New("users not found")
var BusinessErrUsernameTaken = errors.New("the username is already taken")
var BusinessErrEmailTaken = errors.New("the email is already in use")
var BusinessErrUnknownConflict = errors.New("unknown conflict")

const uniqueConstraintViolationCode = "23505"

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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, BusinessErrNoUsers
		}
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, BusinessErrNoUsers
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) DeleteByUuid(uuid uuid.UUID) error {
	result, err := r.db.ExecContext(context.Background(), `DELETE FROM users WHERE uuid = $1`, uuid)
	if err != nil {
		return err
	}

	return ensureSomeRowsAffected(result)
}

func (r *userRepository) PartiallyUpdateByUUID(
	uuid uuid.UUID,
	patch dto.UserPatch,
) error {
	setParts := []string{}
	args := []interface{}{}
	sqlPlaceholderIndex := 1

	if patch.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", sqlPlaceholderIndex))
		args = append(args, patch.Username)
		sqlPlaceholderIndex++
	}

	if patch.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", sqlPlaceholderIndex))
		args = append(args, patch.Email)
		sqlPlaceholderIndex++
	}

	if patch.FullName != nil {
		setParts = append(setParts, fmt.Sprintf("full_name = $%d", sqlPlaceholderIndex))
		if patch.FullName.Value == nil {
			args = append(args, nil)
		} else {
			args = append(args, patch.FullName.Value)
		}
		sqlPlaceholderIndex++
	}

	if len(setParts) == 0 {
		return nil
	}

	args = append(args, uuid)

	// #nosec G201 -- placeholders are still in place
	query := fmt.Sprintf(`UPDATE users SET %s WHERE uuid = $%d`,
		strings.Join(setParts, ", "),
		sqlPlaceholderIndex)

	result, err := r.db.ExecContext(context.Background(), query, args...)
	if err != nil {
		return processConstraintViolations(err)
	}

	return ensureSomeRowsAffected(result)
}

func (r *userRepository) Create(user dto.UserCreate) (*model.User, error) {
	var fullNameValue sql.NullString
	if user.FullName != nil {
		fullNameValue = sql.NullString{
			String: *user.FullName,
			Valid:  true,
		}
	}

	const query = `
        INSERT INTO users (username, email, full_name)
        VALUES ($1, $2, $3)
        RETURNING id, uuid, username, email, full_name
    `

	var createdUser model.User
	err := r.db.QueryRowContext(
		context.Background(),
		query,
		user.Username, user.Email, fullNameValue,
	).Scan(
		&createdUser.ID, &createdUser.UUID, &createdUser.Username, &createdUser.Email, &createdUser.FullName)

	if err != nil {
		return nil, processConstraintViolations(err)
	}

	return &createdUser, nil
}

func processConstraintViolations(err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case uniqueConstraintViolationCode:
			switch pgErr.Constraint {
			case "users_username_key":
				return BusinessErrUsernameTaken
			case "users_email_key":
				return BusinessErrEmailTaken
			default:
				return BusinessErrUnknownConflict
			}
		}
	}

	return err
}

func ensureSomeRowsAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return BusinessErrNoUsers
	}

	return nil
}
