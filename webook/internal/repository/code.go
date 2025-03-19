package repository

import (
	"context"
	"dream/webook/internal/repository/cache"
)

var (
	ErrSetCodeSendTooMany = cache.ErrSetCodeSendTooMany
	ErrVerifyTooManyTimes = cache.ErrVerifyTooManyTimes
)

type CodeRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error)
}

type CacheCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) CodeRepository {
	return &CacheCodeRepository{
		cache: cache,
	}
}

func (repo *CacheCodeRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CacheCodeRepository) Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputcode)
}
