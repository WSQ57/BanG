package service

import (
	"context"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicate = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("用户名或密码不正确")
var ErrUserNotFound = repository.ErrUserNotFound

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	EditProfile(ctx context.Context, u domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type NoramlUserService struct {
	Repo  repository.UserRepository
	redis *redis.Client
}

func NewUserService(repo repository.UserRepository) UserService {
	return &NoramlUserService{
		Repo: repo,
	}
}

func (svc *NoramlUserService) Signup(ctx context.Context, u domain.User) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.Repo.Create(ctx, u)
}

func (svc *NoramlUserService) CacheSignup(ctx context.Context, u domain.User) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	err = svc.Repo.Create(ctx, u)
	if err != nil {
		return err
	}
	// redis 处理 u
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	svc.redis.Set(ctx, fmt.Sprintf("user:profile:%d", u.Id), val, time.Minute*30)

	return err
}

func (svc *NoramlUserService) Login(ctx context.Context, email, password string) (domain.User, error) {

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

func (svc *NoramlUserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.Repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (svc *NoramlUserService) EditProfile(ctx context.Context, u domain.User) error {
	_, err := svc.Repo.FindById(ctx, u.Id)
	if err != nil {
		return repository.ErrUserNotFound
	}
	return svc.Repo.EditProfile(ctx, u)
}

func (svc *NoramlUserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 快路径
	u, err := svc.Repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		// nil 也进来
		return u, err
	}

	// 慢路径
	// 在系统资源不足后，触发降级之后，不执行慢路径
	if ctx.Value("降级") == "true" {
		return domain.User{}, errors.New("系统降级了")
	}

	u = domain.User{
		Phone: phone,
	}
	err = svc.Repo.Create(ctx, u)
	if err != nil {
		return domain.User{}, err
	}
	// 会遇到主从延迟问题
	return svc.Repo.FindByPhone(ctx, phone)

}

func PathDownGrade(ctx context.Context, quick, slow func()) {
	quick()
	if ctx.Value("降级") == "true" {
		return
	}
	slow()
}

func (svc *NoramlUserService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := svc.Repo.FindByWechat(ctx, wechatInfo.OpenID)
	fmt.Println("service--1--", err)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	u = domain.User{
		WechatInfo: wechatInfo,
	}
	err = svc.Repo.Create(ctx, u)
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.Repo.FindByWechat(ctx, wechatInfo.OpenID)
}
