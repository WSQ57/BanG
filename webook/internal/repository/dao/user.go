package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("邮箱or手机号冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {

	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// ms
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictErr uint16 = 1062 // 唯一索引错误吗
		if mysqlErr.Number == uniqueConflictErr {
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error

	return u, err
}

func (dao *UserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error

	return u, err
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (dao *UserDAO) Update(ctx context.Context, u User) error {
	return dao.db.WithContext(ctx).Where("id = ?", u.Id).Updates(&u).Error
}

// User直接对应数据库表结构
// 有些人叫做entity 有些叫做model 有些人叫做PO
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 所有用户唯一
	Email sql.NullString `gorm:"unique"` // 唯一索引，允许有多个空值

	// 如果设置为string `gorm:"unique"` 通过email注册，空字符串会唯一索引冲突
	Phone    sql.NullString `gorm:"unique"`
	Password string

	// 新增字段
	Nickname string `gorm:"type=varchar(128)"`
	Birthday int64
	AboutMe  string `gorm:"type=varchar(4096)"`

	// 创建时间 ms
	Ctime int64
	// 更新时间 ms
	Utime int64
}
