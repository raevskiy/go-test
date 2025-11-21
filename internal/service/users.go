package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	DeleteByUuid(uuid uuid.UUID) error
}

type userService struct {
	repo repository.UserRepository
}

var BusinessErrNoUsers = errors.New("users not found")

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAll() ([]model.User, error) {
	return s.repo.GetAll()
}

func (s *userService) GetByUsername(username string) (*model.User, error) {
	user, err := s.repo.GetByUsername(username)

	return getSingleUser(user, err)
}

func (s *userService) GetByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)

	return getSingleUser(user, err)
}

func (s *userService) DeleteByUuid(uuid uuid.UUID) error {
	err := s.repo.DeleteByUuid(uuid)
	if errors.Is(err, sql.ErrNoRows) {
		return BusinessErrNoUsers
	}

	return err
}

func getSingleUser(user *model.User, err error) (*model.User, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return nil, BusinessErrNoUsers
	}

	return user, err
}

