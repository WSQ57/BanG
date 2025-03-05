package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
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
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error

	return u, err
}

// func (dao *UserDAO) Search(ctx context.Context, u User) error {

// 	var user User
// 	err := dao.db.WithContext(ctx).First(&user, "Email = ?", u.Email).Error
// 	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
// 		const uniqueConflictErr uint16 = 1062 // 唯一索引错误吗
// 		if mysqlErr.Number == uniqueConflictErr {
// 			return ErrUserDuplicateEmail
// 		}
// 	}
// 	return err
// }

// User直接对应数据库表结构
// 有些人叫做entity 有些叫做model 有些人叫做PO
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 所有用户唯一
	Email    string `gorm:"unique"`
	Password string

	// 创建时间 ms
	Ctime int64
	// 更新时间 ms
	Utime int64
}
