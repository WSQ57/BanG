package repository

import (
	"context"
	"dream/webook/internal/repository/cache"
)

var (
	ErrSetCodeSendTooMany = cache.ErrSetCodeSendTooMany
	ErrVerifyTooManyTimes = cache.ErrVerifyTooManyTimes
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(cache *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: cache,
	}
}

func (repo *CodeRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CodeRepository) Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputcode)
}
