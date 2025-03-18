package repository

import (
	"context"
	"database/sql"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository/cache"
	"dream/webook/internal/repository/dao"
	"fmt"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: c,
	}
}

// 没有注册概念
func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	return r.dao.Insert(ctx, r.domainToEntity(u))
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
	return r.entityToDomain(user), nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	user, err := r.dao.FindByPhone(context.Background(), phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	u, err := r.cache.Get(ctx, id)

	if err == nil {
		return u, nil
	}
	// 需要从数据库中加载
	if err == cache.ErrKeyNotExist {
		ue, err := r.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}

		u = r.entityToDomain(ue)
		_ = r.cache.Set(ctx, u) // 忽略错误
		return u, nil
	}

	// 如果不等于nil还不等于errkey 有可能是redis整个崩溃 也有可能是偶发性事故
	// err = io.EOF
	// 如果决定加载数据库 要考虑redis崩溃问题，保护数据库——限流数据库
	// 不加载的话 用户体验差

	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = r.entityToDomain(ue)
	err = r.cache.Set(ctx, u)
	go func() { // 异步来加速，缓存本来就有数据一致性问题
		if err != nil {
			// 缓存失败，不阻塞业务逻辑
			// 继续返回数据库查询的结果
			// 打日志做监控 此时，查询失败、存也失败
			fmt.Println("缓存失败")
		}
	}()
	return u, nil
}

func (r *UserRepository) entityToDomain(ud dao.User) domain.User {
	return domain.User{
		Id:       ud.Id,
		Email:    ud.Email.String,
		Password: ud.Password,
		Phone:    ud.Phone.String,
		Nickname: ud.Nickname,
		Birthday: time.Unix(ud.Birthday, 0),
		Ctime:    time.Unix(ud.Ctime, 0),
		AboutMe:  ud.AboutMe,
	}
}

func (r *UserRepository) domainToEntity(ud domain.User) dao.User {
	return dao.User{
		Id: ud.Id,
		Email: sql.NullString{
			String: ud.Email,
			Valid:  ud.Email != "",
		},
		Phone: sql.NullString{
			String: ud.Phone,
			Valid:  ud.Phone != "",
		},
		Password: ud.Password,
		Nickname: ud.Nickname,
		Birthday: ud.Birthday.Unix(),
		Ctime:    ud.Ctime.Unix(),
		AboutMe:  ud.AboutMe,
	}
}
