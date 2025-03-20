package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
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

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
	key(biz, phone string) string
}

type RedisCodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (cache *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
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

func (cache *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
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

func (cache *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

// LocalLRUCodeCache :基于LocalLRUCodeCache的本地缓存实现
type LocalLRUCodeCache struct {
	expire time.Duration // 过期时间
	mu     sync.Mutex
	cache  *expirable.LRU[string, any]
}

func NewLocalLRUCodeCache(exp time.Duration) CodeCache {
	cache := expirable.NewLRU[string, any](200, nil, exp)
	return &LocalLRUCodeCache{
		expire: exp,
		mu:     sync.Mutex{},
		cache:  cache,
	}
}

func (l *LocalLRUCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := l.key(biz, phone)          // 存储验证码的key
	keyCount := key + ":count"        // 存储验证码1分钟内验证了几次的key
	keySetTime := key + ":keySetTime" // 存储验证码存入时间的key
	//cache := expirable.NewLRU[string, any](200, nil, l.expire)

	l.mu.Lock()
	defer l.mu.Unlock()
	// 如果key不存在则写入
	existOk := l.cache.Contains(key)       // key 是否存在
	setTimeOk := false                     // 在有效期内是否是大于一分钟校验的
	setTime, ok := l.cache.Get(keySetTime) // 设置的写入时间是否存在
	if ok {
		setTimeValue, _ := setTime.(time.Time)
		_ = setTimeValue.Add(time.Minute * 10)
		if time.Now().After(setTimeValue.Add(time.Minute * 1)) {
			setTimeOk = true
		} else {
			return ErrSetCodeSendTooMany
		}
	}

	if !existOk || setTimeOk {
		l.cache.Add(key, code)
		l.cache.Add(keyCount, 3)
		l.cache.Add(keySetTime, time.Now())
	}

	return nil
}

func (l *LocalLRUCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	key := l.key(biz, phone)   // 存储验证码的key
	keyCount := key + ":count" // 存储验证码1分钟内验证了几次的key

	l.mu.Lock()
	defer l.mu.Unlock()

	count, ok := l.cache.Get(keyCount) // 验证码验证次数
	if !ok {
		return false, ErrUnknownForCode
	}
	if count.(int) <= 0 {
		return false, ErrVerifyTooManyTimes
	}
	code, ok := l.cache.Get(key) // 存储的code
	if !ok {
		return false, ErrUnknownForCode
	}
	// 用户输对
	if code.(string) == inputCode {
		l.cache.Add(keyCount, 1)
		return true, nil
	}
	// 用户输错
	l.cache.Add(keyCount, count.(int)-1)
	return false, nil
}

func (c *LocalLRUCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
