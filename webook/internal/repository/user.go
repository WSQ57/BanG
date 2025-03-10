package repository

import (
	"context"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

// 没有注册概念
func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) EditProfile(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		AboutMe:  u.AboutMe,
		Birthday: u.Birthday.Unix(),
	})
}

func (r *UserRepository) FindByEmail(ctx context.Context, Email string) (domain.User, error) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	user, err := r.dao.FindByEmail(context.Background(), Email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
		Nickname: user.Nickname,
		Birthday: time.Unix(user.Birthday, 0),
		Ctime:    time.Unix(user.Ctime, 0),
		AboutMe:  user.AboutMe,
	}, nil
}
func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: time.Unix(u.Birthday, 0),
		Ctime:    time.Unix(u.Ctime, 0),
		AboutMe:  u.AboutMe,
	}, nil
}
