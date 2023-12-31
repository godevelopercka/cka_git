package repository

import (
	"context"
	"database/sql"
	"practice/webook/internal/domain"
	"practice/webook/internal/repository/cache"
	"practice/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFount
)

//写法二：var ErrUserDuplicateEmailV1 = fmt.Errorf("%w 邮箱冲突", dao.ErrUserDuplicateEmail)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByWechat(ctx context.Context, OpenID string) (domain.User, error)
}

type CachedUserRepository struct { // 不要在这里面初始化
	dao   dao.UserDAO
	cache cache.UserCache
}

// func NewUserRepository(dao *dao.GORMUserDAO, c *cache.RedisUserCache) *UserRepository { // 通过传参初始化
func NewUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository { // 通过传参初始化
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *CachedUserRepository) FindByWechat(ctx context.Context, OpenID string) (domain.User, error) {
	u, err := r.dao.FindByWechat(ctx, OpenID)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToDomain(u))
}

func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 里面找
	// 再从 dao 里面找
	// 找到了回写 cache
	u, err := r.cache.Get(ctx, id)
	if err != nil {
		// 必然是有数据
		return u, nil
	}
	// 没这个数据
	//if err == cache.ErrKeyNotExist {
	// 去数据库里面加载
	//}

	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = r.entityToDomain(ue)

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 我这里怎么办？
			// 打日志，做监控
			//return domain.User{}, err
		}
	}()

	return u, err

	// 这里怎么办？ err = io.EOF
	// 要不要去数据库加载？
	// 看起来我不应该加载
	// 看起来我好像也要加载？

	// 选加载 —— 做好兜底，万一 Redis 真的崩了。你要保护住你的数据库
	// 我数据库限流呀！

	// 选不加载 —— 用户体验差一点

	// 缓存里面有数据
	// 缓存里面没有数据
	// 缓存出错了，你也不知道有没有数据
}

func (r *CachedUserRepository) domainToDomain(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			// 我确实有手机号
			Valid: u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		WechatOpenID: sql.NullString{
			String: u.WechatInfo.OpenID,
			Valid:  u.WechatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: u.WechatInfo.UnionID,
			Valid:  u.WechatInfo.UnionID != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		WechatInfo: domain.WechatInfo{
			UnionID: u.WechatUnionID.String,
			OpenID:  u.WechatOpenID.String,
		},
		Ctime: time.UnixMilli(u.Ctime),
	}
}
