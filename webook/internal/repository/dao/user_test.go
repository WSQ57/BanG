package dao

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCase := []struct {
		name string
		ctx  context.Context
		user User

		mock func(t *testing.T) *sql.DB

		wantErr error
		wantId  int64
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()

				res := sqlmock.NewResult(3, 1)
				// 输入正则表达式
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
				Password: "123",
			},
			wantId:  3,
			wantErr: nil,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {

			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn: tc.mock(t),
				// SELECT VERSION
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				SkipDefaultTransaction: true,
				DisableAutomaticPing:   true,
			})
			assert.NoError(t, err)
			d := NewUserDAO(db)
			err = d.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
