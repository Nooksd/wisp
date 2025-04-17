package service

import (
	"context"
	"errors"
	"fmt"

	"wisp/src/model"
	"wisp/src/repository"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *repository.UserRepo
	validator *validator.Validate
}

func NewUserService(r *repository.UserRepo) *UserService {
	return &UserService{repo: r, validator: validator.New()}
}

func (s *UserService) CheckAvailability(ctx context.Context, email, userID string) (bool, error) {
	return s.repo.IsAvailable(ctx, email, userID)
}

func (s *UserService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {

	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validação falhou: %w", err)
	}

	avail, err := s.repo.IsAvailable(ctx, req.Email, req.UserID)
	if err != nil {
		return nil, err
	}
	if !avail {
		return nil, errors.New("email ou userId já cadastrado")
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		UserID:       req.UserID,
		PasswordHash: string(hash),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) GetByUserID(ctx context.Context, userID string) (*model.User, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *UserService) ListUsers(ctx context.Context, page, limit int, q, sortField string, desc bool) ([]model.User, int64, error) {
	return s.repo.List(ctx, page, limit, q, sortField, desc)
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, upd map[string]interface{}) error {
	return s.repo.UpdateByUserID(ctx, userID, upd)
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.DeleteByUserID(ctx, userID)
}
