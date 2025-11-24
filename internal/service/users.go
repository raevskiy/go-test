package service

import (
	"cruder/internal/controller/dto"
	"cruder/internal/model"
	"cruder/internal/repository"
	"errors"
	"github.com/google/uuid"
	"log/slog"
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	DeleteByUuid(uuid uuid.UUID) error
	PartiallyUpdateByUuid(uuid uuid.UUID, patch dto.UserPatch) error
	Create(user dto.UserCreate) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

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
	return s.repo.DeleteByUuid(uuid)
}

func (s *userService) PartiallyUpdateByUuid(uuid uuid.UUID, patch dto.UserPatch) error {
	return s.repo.PartiallyUpdateByUUID(uuid, patch)
}

func (s *userService) Create(user dto.UserCreate) (*model.User, error) {
	return s.repo.Create(user)
}

func getSingleUser(user *model.User, err error) (*model.User, error) {
	if errors.Is(err, repository.BusinessErrNoUsers) {
		slog.Warn("users not found")
	}

	return user, err
}

