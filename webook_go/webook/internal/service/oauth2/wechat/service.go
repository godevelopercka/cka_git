package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"webook_go/webook/internal/domain"
	"webook_go/webook/pkg/logger"
)

var redirectURI = url.PathEscape("https://meoying.com/oauth2/wechat/callback")

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appId     string
	appSecret string
	client    *http.Client
	l         logger.LoggerV1
}

func NewService(appId string, appSecret string, l logger.LoggerV1) Service {
	return &service{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
		l:         l,
	}
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil) // 这种不会产生复制
	if err != nil {
		return domain.WechatInfo{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	// 读取响应，只读一遍
	decoder := json.NewDecoder(resp.Body)
	var res Result
	err = decoder.Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("微信返回错误响应，错误码: %d, 错误信息: %s", res.ErrCode, res.ErrMsg)
	}

	zap.L().Info("调用微信，拿到用户信息", zap.String("unionID", res.UnionID),
		zap.String("openID", res.OpenID))

	return domain.WechatInfo{
		OpenID:  res.OpenID,
		UnionID: res.UnionID,
	}, nil
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpireIn     string `json:"expire_in"`
	RefreshToken string `json:"refresh_token"`

	OpenID  string `json:"openid"`
	Scope   string `json:"scope"`
	UnionID string `json:"unionid"`
}
