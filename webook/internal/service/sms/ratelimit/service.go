package ratelimit

import (
	"context"
	"dream/webook/internal/service/sms"
	"dream/webook/pkg/ratelimit"
	"fmt"
)

var errLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func (r RatelimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, "send:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候，
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		// 包一下这个错误
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		return errLimited
	}
	// 你这里加一些代码，新特性
	err = r.svc.Send(ctx, tpl, args, numbers...)
	// 你在这里也可以加一些代码，新特性
	return err
}

func NewRateLimitSMSServcie(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}
