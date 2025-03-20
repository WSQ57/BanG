package web

import (
	"bytes"
	"context"
	"dream/webook/internal/domain"
	"dream/webook/internal/repository"
	repomocks "dream/webook/internal/repository/mocks"
	"dream/webook/internal/service"
	svcmocks "dream/webook/internal/service/mocks"
	ijwt "dream/webook/internal/web/jwt"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestEncrypt(t *testing.T) {
	password := "helloworld"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

func TestNil(t *testing.T) {
	testTypeAssert(nil)
}

func testTypeAssert(c any) {
	_, ok := c.(*ijwt.UserClaims)
	println(ok)
}

func TestUserHandler_Signup(t *testing.T) {
	testCases := []struct {
		name string

		// 定义一个mock
		mock func(ctrl *gomock.Controller) service.UserService

		// 请求的body
		reqBody string

		// 期待的状态码
		wantCode int
		wantBody string
	}{
		{
			name: "参数不对，bind失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody:  `{"email": "123@qq.com","password": "helloworld123","confirm": "helloworld123",}`,
			wantCode: http.StatusBadRequest,
			wantBody: "系统错误",
		},
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "1234@qq.com",
					Password: "helloworld123",
				}).Return(nil)
				return usersvc
			},
			reqBody: `
					{
						"email": "1234@qq.com",
						"password": "helloworld123",
						"confirmPassword": "helloworld123"
					}			
					`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "邮箱校验失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
					{
						"email": "123",
						"password": "helloworld123",
                        "confirmPassword": "helloworld123"
					}			
					`,
			wantCode: http.StatusOK,
			wantBody: "邮箱格式不正确",
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "11",
                        "confirmPassword": "11"
					}			
					`,
			wantCode: http.StatusOK,
			wantBody: "密码小于9位，且由数字字母组成",
		},
		{
			name: "两次密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "helloworld123",
		                "confirmPassword": "helloworld1234"
					}
					`,
			wantCode: http.StatusOK,
			wantBody: "两次密码不一致",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "helloworld123",
				}).Return(service.ErrUserDuplicate)
				return usersvc
			},
			reqBody: `
						{
							"email": "123@qq.com",
							"password": "helloworld123",
							"confirmPassword": "helloworld123"
						}
						`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			userHandler := NewUserHandler(tc.mock(ctrl), nil) // 注册这里用不到codesvc
			userHandler.RegisterRoutes(server.Group("/users"))

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			// 定义一个响应接收器
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req) // 这个就是HTTP进入GIN框架的入口
			// 当这样调用的时候，GIN就会处理这个请求，并将响应写回到resp里

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usersvc := svcmocks.NewMockUserService(ctrl)

	// 预期有调用
	usersvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(errors.New("mock error")) // return与Signup返回值类型要一致

	err := usersvc.Signup(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}

func TestUserHandler_Login(t *testing.T) {
	// 做成测试用例用到的事件
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		ctx      context.Context
		email    string
		password string

		wantUser domain.User
		wantErr  error
	}{
		{
			name:     "登录成功",
			email:    "1234@qq.com",
			password: "helloworld123",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "1234@qq.com").Return(domain.User{
					Email:    "1234@qq.com",
					Password: "$2a$10$kNDDNA.mFPkmZ3fF6oQ/pev6CZpvNpho2q3Hp7dtHQzt759a7ITCK",
					Ctime:    now,
				}, nil)
				return userRepo
			},
			wantErr: nil,
			wantUser: domain.User{
				Email:    "1234@qq.com",
				Password: "$2a$10$kNDDNA.mFPkmZ3fF6oQ/pev6CZpvNpho2q3Hp7dtHQzt759a7ITCK",
				Ctime:    now,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := service.NewUserService(tc.mock(ctrl))
			u, err := svc.Login(tc.ctx, tc.email, tc.password)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)

		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("helloworld123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
