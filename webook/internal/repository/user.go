package repository

import (
	"context"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository/dao"
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
	}, nil
}
func (r *UserRepository) FindById(int64) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache

}
