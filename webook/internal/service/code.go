package service

import (
	"context"
	"dream/webook/internal/repository"
	"dream/webook/internal/service/sms"
	"fmt"
	"math/rand"
)

const codeTplId = "1877556"

var (
	ErrSetCodeSendTooMany = repository.ErrSetCodeSendTooMany
)

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
	// tplId  string
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
		// tplId:  codeTplId,
	}
}

// biz区别业务场景
func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	// 生成验证码
	code := svc.generateCode()
	// 塞进去redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)

	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz string, phone string, code string) (bool, error) {
	//phonecode:$biz:$phone
	return svc.repo.Verify(ctx, biz, phone, code)
}

// func (svc *CodeService) VerifyV1(ctx context.Context, biz string, phone string, code string) error {

// }

func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000)
	// 不够6位，补0
	return fmt.Sprintf("%06d", num)
}
