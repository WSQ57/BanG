package wechat

import (
	"context"
	"dream/webook/internal/domain"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// PathEscape 会将字符串转义出来，以便将其安全地放置在 URL 路径段中。
var redirectURI = url.PathEscape("https://meoying.com/oauth2/wechat/callback")

type Service interface {
	AuthUrl(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appId     string
	appSecret string
	client    *http.Client
}

func NewService(appId string, appSecret string) Service {
	return &service{
		appId:     appId,
		appSecret: appSecret,
		// 依赖注入，但是没完全注入
		client: http.DefaultClient,
	}
}

func (s service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	// 构造通过code获取UnionID 的url
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)
	// 构造请求
	req, _ := http.NewRequestWithContext(ctx, "GET", target, nil)
	// 发送请求 获取响应
	resp, err := s.client.Do(req)

	//会产生复制，效果很差
	// req = req.WithContext(ctx)

	if err != nil {
		return domain.WechatInfo{}, err
	}
	var result Result
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil || result.ErrCode != 0 {
		return domain.WechatInfo{}, err
	}
	return domain.WechatInfo{
		UnionID: result.UnionID,
		OpenID:  result.OpenID,
	}, nil
}

func (s service) AuthUrl(ctx context.Context, state string) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	// state := "login"
	// state := uuid.NewV4()

	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	OpenID  string `json:"openid"`
	Scope   string `json:"scope"`
	UnionID string `json:"unionid"`
}
