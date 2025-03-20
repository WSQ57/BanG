package tencent

import (
	"context"
	"fmt"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
	// limiter  ratelimit.Limiter
}

func NewService(client *sms.Client, appId, signName string) *Service {
	return &Service{
		appId:    ekit.ToPtr(appId),
		signName: ekit.ToPtr(signName),
		client:   client,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	// limited, err := s.limiter.Limit(ctx, "sms:tencent")
	// if err != nil {
	// 	// 系统错误
	// 	// 可以限流，保守策略，下游很坑的时候
	// 	// 可以不限，下游很强，业务可用性要求高，尽量容错策略
	// 	// 包一下错误
	// 	return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	// }
	// if limited {
	// 	// 限流了
	// 	return fmt.Errorf("短信服务限流了")
	// }

	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &tplId
	req.PhoneNumberSet = s.toStringPtrSlice(number)
	req.TemplateParamSet = s.toStringPtrSlice(args)
	resp, err := s.client.SendSms(req)
	if err != nil {
		// TODO: 错误处理
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *status.Code != "Ok" {
			return fmt.Errorf("短信发送失败 %s, %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map(src, func(idx int, src string) *string {
		return &src
	})
}
