package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"database/sql"
	"errors"
	"log"
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

var ErrNoUsers = errors.New("service: no user matching the search criteria found")

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

func getSingleUser(user *model.User, err error) (*model.User, error) {
	if errors.Is(err, sql.ErrNoRows) {
		log.Println("users not found")

		return nil, ErrNoUsers
	}

	return user, err
}

