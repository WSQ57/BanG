package Integration

import (
	"bytes"
	"context"
	"dream/webook/internal/web"
	"dream/webook/ioc"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseHandler_SendLoginSMSCode(t *testing.T) {
	server := initWebServer()
	rdb := ioc.InitRedis()
	testCase := []struct {
		name string

		phone string
		// 准备测试数据
		before func(t *testing.T)
		// 验证数据，清理数据
		after func(t *testing.T)

		reqBody  string
		wantCode int
		wantBody web.Result
	}{
		{
			name:   "发送成功",
			phone:  "13588888888",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				// 发送成功后，要清理redis中的数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				val, err := rdb.GetDel(ctx, "phone_code:login:13588888888").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
				{
					"phone":"13588888888"
				}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name:   "手机号输入有误",
			phone:  "13588888888",
			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},
			reqBody: `
				{
					"tel_phone":"13588888888",
				}
			`,
			wantCode: 400,
			wantBody: web.Result{
				Code: 4,
				Msg:  "输入有误",
			},
		},
		{
			name:  "发送太频繁",
			phone: "13588888888",
			before: func(t *testing.T) {
				// 发送之前，先往Redis中塞一条数据，模拟30s内刚刚发过一次
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.Set(ctx, "phone_code:login:13588888888", "123456", time.Minute*9+time.Second*30)
				cancel()
			},
			after: func(t *testing.T) {
				// 发送成功后，要清理redis中的数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				val, err := rdb.GetDel(ctx, "phone_code:login:13588888888").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
				{
					"phone":"13588888888"
				}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送太频繁，请稍后再试",
			},
		},
		{
			name:  "系统错误",
			phone: "13588888888",
			before: func(t *testing.T) {
				// 提前塞一个验证码，但是没有过期时间
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				_, err := rdb.Set(ctx, "phone_code:login:13588888888", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 发送成功后，要清理redis中的数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				val, err := rdb.GetDel(ctx, "phone_code:login:13588888888").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
				{
					"phone":"13588888888"
				}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			// 准备数据
			tc.before(t)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			// 定义一个响应接收器
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req) // 这个就是HTTP进入GIN框架的入口
			// 当这样调用的时候，GIN就会处理这个请求，并将响应写回到resp里

			// 验证数据
			assert.Equal(t, resp.Code, tc.wantCode)
			if resp.Code != 200 {
				return
			}
			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, webRes, tc.wantBody)

			tc.after(t)
		})
	}

}

func TestUserHandler_LoginJWT(t *testing.T) {
	server := initWebServer()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "登录成功",
			reqBody: `
				{
					"email":"a@qq.com",
                    "password":"admin123!"
				}
			`,
			before:   func(t *testing.T) {},
			after:    func(t *testing.T) {},
			wantCode: 200,
			wantBody: "登录成功",
		},
		{
			name: "用户名或密码不对",
			reqBody: `
				{
					"email":"a@qq.com",
                    "password":"admin1234!"
				}
			`,
			before:   func(t *testing.T) {},
			after:    func(t *testing.T) {},
			wantCode: 200,
			wantBody: "用户名或密码不对",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			// 定义一个响应接收器
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req) // 这个就是HTTP进入GIN框架的入口
			// 验证
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, tc.wantBody)
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	server := initWebServer()
	rdb := ioc.InitRedis()
	testcases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		reqBody  string
		wantCode int
		wantBody web.Result
	}{
		{
			name: "验证码校验通过",
			before: func(t *testing.T) {
				// 提前把验证码存入redis
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.Set(ctx, "phone_code:login:13588888888", "123456", time.Minute*10)
				rdb.Set(ctx, "phone_code:login:13588888888:cnt", 3, time.Minute*10)
				cancel()
			},
			after: func(t *testing.T) {
				// 清理数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.GetDel(ctx, "phone_code:login:13588888888")
				rdb.GetDel(ctx, "phone_code:login:13588888888:cnt")
				cancel()
			},

			reqBody: `
			{
				"phone":"13588888888",
				"code":"123456"
			}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "验证码校验通过",
			},
		},
		{
			name: "验证码有误",
			before: func(t *testing.T) {
				// 提前把验证码存入redis
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.Set(ctx, "phone_code:login:13588888888", "123456", time.Minute*10)
				rdb.Set(ctx, "phone_code:login:13588888888:cnt", 3, time.Minute*10)
				cancel()
			},
			after: func(t *testing.T) {
				// 清理数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.GetDel(ctx, "phone_code:login:13588888888")
				rdb.GetDel(ctx, "phone_code:login:13588888888:cnt")
				cancel()
			},

			reqBody: `
			{
				"phone":"13588888888",
				"code":"123457"
			}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 4,
				Msg:  "验证码有误",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				// 提前把验证码存入redis
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.Set(ctx, "phone_code:login:13588888888", "123456", time.Minute*10)
				rdb.Set(ctx, "phone_code:login:13588888888:cnt", -1, time.Minute*10)
				cancel()
			},
			after: func(t *testing.T) {
				// 清理数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				rdb.GetDel(ctx, "phone_code:login:13588888888")
				rdb.GetDel(ctx, "phone_code:login:13588888888:cnt")
				cancel()
			},

			reqBody: `
			{
				"phone":"13588888888",
				"code":"123456"
			}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			// 定义一个响应接收器S
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req) //进入GIN框架的入口

			// 清理数据
			tc.after(t)
			// 验证
			assert.Equal(t, tc.wantCode, resp.Code)
			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)

		})
	}

}
