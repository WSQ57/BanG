package repository

import (
	"context"
	"database/sql"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository/cache"
	cachemocks "dream/webook/internal/repository/cache/mocks"
	"dream/webook/internal/repository/dao"
	daomocks "dream/webook/internal/repository/dao/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCacheUserRepository_FindById(t *testing.T) {
	now := time.UnixMilli(time.Now().UnixMilli())
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx  context.Context
		id   int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中,查询成功",

			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)

				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{
					Id:       123,
					Nickname: "jane",
					Email:    sql.NullString{String: "jane@qq.com", Valid: true},
					Phone:    sql.NullString{String: "12314141", Valid: true},
					Birthday: now.UnixMilli(),
					Ctime:    now.UnixMilli(),
					Utime:    now.UnixMilli(),
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Nickname: "jane",
					Email:    "jane@qq.com",
					Phone:    "12314141",
					Birthday: now,
					Ctime:    now,
				}).Return(nil)
				return d, c
			},

			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Nickname: "jane",
				Email:    "jane@qq.com",
				Phone:    "12314141",
				Birthday: now,
				Ctime:    now,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewUserRepository(tc.mock(ctrl))
			u, err := repo.FindById(tc.ctx, tc.id)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}
