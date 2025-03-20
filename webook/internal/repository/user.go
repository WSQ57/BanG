package repository

import (
	"context"
	"database/sql"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository/cache"
	"dream/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	EditProfile(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, Email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
	// entityToDomain(ud dao.User) domain.User
	// domainToEntity(ud domain.User) dao.User
}

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: c,
	}
}

// 没有注册概念
func (r *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CacheUserRepository) EditProfile(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		AboutMe:  u.AboutMe,
		Birthday: u.Birthday.Unix(),
	})
}

func (r *CacheUserRepository) FindByEmail(ctx context.Context, Email string) (domain.User, error) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	user, err := r.dao.FindByEmail(ctx, Email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache
	user, err := r.dao.FindByPhone(context.Background(), phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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

	_ = r.cache.Set(ctx, u)

	// go func() { // 异步来加速，缓存本来就有数据一致性问题
	// 	err = r.cache.Set(ctx, u)
	// 	if err != nil {
	// 		// 缓存失败，不阻塞业务逻辑
	// 		// 继续返回数据库查询的结果
	// 		// 打日志做监控 此时，查询失败、存也失败
	// 		fmt.Println("缓存失败")
	// 	}
	// }()
	return u, nil
}

func (r *CacheUserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	u, err := r.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CacheUserRepository) entityToDomain(ud dao.User) domain.User {
	return domain.User{
		Id:       ud.Id,
		Email:    ud.Email.String,
		Password: ud.Password,
		Phone:    ud.Phone.String,
		Nickname: ud.Nickname,
		Birthday: time.UnixMilli(ud.Birthday),
		Ctime:    time.UnixMilli(ud.Ctime),
		AboutMe:  ud.AboutMe,
	}
}

func (r *CacheUserRepository) domainToEntity(ud domain.User) dao.User {
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
		Birthday: ud.Birthday.UnixMilli(),
		Ctime:    ud.Ctime.UnixMilli(),
		AboutMe:  ud.AboutMe,
	}
}
