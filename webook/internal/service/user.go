package service

import (
	"context"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
var ErrInvalidUserOrPassword = errors.New("用户名或密码不正确")
var ErrUserNotFound = repository.ErrUserNotFound

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.Repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {

	u, err := svc.Repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err != nil {
		// debug
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.Repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (svc *UserService) EditProfile(ctx context.Context, u domain.User) error {
	_, err := svc.Repo.FindById(ctx, u.Id)
	if err != nil {
		return repository.ErrUserNotFound
	}
	return svc.Repo.EditProfile(ctx, u)
}
