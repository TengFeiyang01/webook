package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
	"webook/webook/pkg/logger"
)

var (
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("邮箱或密码不对")
)

type userService struct {
	repo repository.UserRepository
	//logger *zap.Logger
	l logger.LoggerV1
}

func NewUserService(repo repository.UserRepository, l logger.LoggerV1) UserService {
	return &userService{
		repo: repo,
		l:    l,
	}
}

//func NewUserServiceV1(repo repository.UserRepository, l *zap.Logger) UserService {
//	return &userService{
//		repo:   repo,
//		logger: zap.L(),
//	}
//}

// SignUp 注册
func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 存起来
	if err := svc.repo.Create(ctx, u); err != nil {
		return err
	}
	return nil
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 这时候呢, 这个地方怎么做
	u, err := svc.repo.FindByPhone(ctx, phone)
	// 要判断有没有这个用户
	// 这个叫做快路径
	if !errors.Is(err, repository.ErrUserNotFound) {
		// nil、不为ErrUserNotFound 都会进来这里
		return u, err
	}
	// 这里，把 phone 脱敏之后打出来
	svc.l.Info("uses not register", logger.String("phone", phone))

	// 这个叫做慢路径
	// 明确知道, 没有这个用户
	u = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	// 这时候呢, 这个地方怎么做
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	// 要判断有没有这个用户
	// 这个叫做快路径
	if !errors.Is(err, repository.ErrUserNotFound) {
		// nil、不为ErrUserNotFound 都会进来这里
		return u, err
	}
	u = domain.User{
		WechatInfo: info,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	// 先从缓存找
	return svc.repo.FindById(ctx, id)
}

func (svc *userService) UpdateById(ctx context.Context, u domain.User) error {
	return svc.repo.UpdateById(ctx, u)
}

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error {
	return svc.repo.UpdateById(ctx, u)
}
