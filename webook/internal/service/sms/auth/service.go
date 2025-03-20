package auth

import (
	"context"
	"dream/webook/internal/service/sms"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key string
}

// Send 发送，biz必须是线下申请的代表业务方的token
func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {

	var tc Claims

	// 解析成功说明是对应的业务方
	token, err := jwt.ParseWithClaims(biz, &tc, func(t *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token不合法")
	}
	return s.svc.Send(ctx, biz, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	tplId int64
}
