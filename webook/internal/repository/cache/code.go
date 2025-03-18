package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrSetCodeSendTooMany = errors.New("验证码发送太频繁")
	ErrVerifyTooManyTimes = errors.New("验证码错误次数太多")
	ErrUnknownForCode     = errors.New("未知错误")
)

// 编译器在编译时，会把set_code.lua文件嵌入到程序中
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

func (cache *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cache.client.Eval(ctx, luaSetCode, []string{cache.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0: // 毫无问题
		return nil
	case -1: // 发送太频繁
		return ErrSetCodeSendTooMany
	default:
		// 系统错误
		return errors.New("系统错误")
	}
}

func (cache *CodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := cache.client.Eval(ctx, luaVerifyCode, []string{cache.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0: // 毫无问题
		return true, nil
	case -1:
		//
		return false, ErrVerifyTooManyTimes
	case -2:
		return false, nil
	default:
		// 系统错误
		return false, ErrUnknownForCode
	}
}

func (cache *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
