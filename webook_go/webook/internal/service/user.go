package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook_go/webook/internal/domain"
	"webook_go/webook/internal/repository"
	"webook_go/webook/internal/repository/cache"
	"webook_go/webook/pkg/logger"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	Edit(ctx context.Context, id int64, Nickname, Birthday, AboutMe string) (domain.User, error)
	FindOrCreate(ctx context.Context, Phone string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo  repository.UserRepository
	cache cache.CodeRedisCache
	l     logger.LoggerV1
}

func NewUserService(repo repository.UserRepository, c cache.CodeRedisCache, l logger.LoggerV1) UserService {
	return &userService{
		repo:  repo,
		cache: c,
		l:     l,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error { // 调用下层 repository 的方法
	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是，存起来
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) Edit(ctx context.Context, id int64, Nickname, Birthday, AboutMe string) (domain.User, error) {
	u, err := svc.repo.Edit(ctx, id, Nickname, Birthday, AboutMe)
	if err != nil {
		return domain.User{}, errors.New("请注册账号")
	}
	return u, nil
}

func (svc *userService) FindOrCreate(ctx context.Context, Phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, Phone)
	// 要判断，是否有这个用户
	if err != repository.ErrUserNotFound {
		// 绝大部分请求会进来这里，因为很多用户都注册过了
		// 这个叫做快路径
		// nil 会进来这里
		// 不为 ErrUserNotFound 的也会进来这里
		return u, err
	}
	// 这里，把 phone 脱敏之后打出来
	//zap.L().Info("用户未注册", zap.String("phone", Phone))
	svc.l.Info("用户未注册", logger.String("phone", Phone))
	// 在系统资源不足，触发降级之后，不执行慢路径了。相当于优先服务已经注册的用户
	if ctx.Value("降级") == "true" {
		return domain.User{}, errors.New("系统降级了")
	}
	// 这个叫做慢路径
	// 你明确知道，没有这个用户
	u = domain.User{
		Phone: Phone,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil || err != repository.ErrUserDuplicate {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, Phone)
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	// 要判断，是否有这个用户
	if err != repository.ErrUserNotFound {
		return u, err
	}

	if ctx.Value("降级") == "true" {
		return domain.User{}, errors.New("系统降级了")
	}
	u = domain.User{
		WechatInfo: info,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil || err != repository.ErrUserDuplicate {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	// 个人版-无缓存
	//u, err := svc.repo.Profile(ctx, id)
	//if err != nil {
	//	return domain.User{}, errors.New("暂无提交个人信息")
	//}
	//return u, nil

	// 大明版-先从缓存查
	u, err := svc.repo.FindById(ctx, id)
	return u, err
}
