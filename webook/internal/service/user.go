package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

var (
	ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrEmail = errors.New("邮箱或密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, err
	}

	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvalidUserOrEmail
	}
	return u, nil
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 你要考虑加密放在哪里
	// 然后就是存起来
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hash)

	return svc.repo.Create(ctx, u)
}