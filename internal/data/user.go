package data

import (
	"context"
	"time"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/sysuser"
	"github.com/weetime/agent-matrix/internal/data/ent/sysusertoken"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

type userTokenRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo 初始化UserRepo
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  kit.LogHelper(logger),
	}
}

// NewUserTokenRepo 初始化UserTokenRepo
func NewUserTokenRepo(data *Data, logger log.Logger) biz.UserTokenRepo {
	return &userTokenRepo{
		data: data,
		log:  kit.LogHelper(logger),
	}
}

// GetByUsername 根据用户名获取用户
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	user, err := r.data.db.SysUser.Query().
		Where(sysuser.Username(username)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizUser biz.User
	if err := copier.Copy(&bizUser, user); err != nil {
		return nil, err
	}
	return &bizUser, nil
}

// GetByUserId 根据用户ID获取用户
func (r *userRepo) GetByUserId(ctx context.Context, userId int64) (*biz.User, error) {
	user, err := r.data.db.SysUser.Get(ctx, userId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizUser biz.User
	if err := copier.Copy(&bizUser, user); err != nil {
		return nil, err
	}
	return &bizUser, nil
}

// Save 保存用户
func (r *userRepo) Save(ctx context.Context, user *biz.User) error {
	create := r.data.db.SysUser.Create().
		SetUsername(user.Username).
		SetPassword(user.Password).
		SetSuperAdmin(user.SuperAdmin).
		SetStatus(user.Status)

	if user.Creator > 0 {
		create.SetCreator(user.Creator)
	}
	if user.Updater > 0 {
		create.SetUpdater(user.Updater)
	}
	if !user.CreateDate.IsZero() {
		create.SetCreateDate(user.CreateDate)
	} else {
		create.SetCreateDate(time.Now())
	}

	return create.Exec(ctx)
}

// Update 更新用户
func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	update := r.data.db.SysUser.Update().
		Where(sysuser.ID(user.ID))

	if user.Username != "" {
		update.SetUsername(user.Username)
	}
	if user.Password != "" {
		update.SetPassword(user.Password)
	}
	if user.SuperAdmin >= 0 {
		update.SetSuperAdmin(user.SuperAdmin)
	}
	if user.Status >= 0 {
		update.SetStatus(user.Status)
	}
	if user.Updater > 0 {
		update.SetUpdater(user.Updater)
	}

	return update.Exec(ctx)
}

// ChangePassword 修改密码
func (r *userRepo) ChangePassword(ctx context.Context, userId int64, newPassword string) error {
	return r.data.db.SysUser.Update().
		Where(sysuser.ID(userId)).
		SetPassword(newPassword).
		Exec(ctx)
}

// ChangePasswordDirectly 直接修改密码（不验证旧密码）
func (r *userRepo) ChangePasswordDirectly(ctx context.Context, userId int64, newPassword string) error {
	return r.data.db.SysUser.Update().
		Where(sysuser.ID(userId)).
		SetPassword(newPassword).
		Exec(ctx)
}

// GetUserCount 获取用户总数
func (r *userRepo) GetUserCount(ctx context.Context) (int64, error) {
	count, err := r.data.db.SysUser.Query().Count(ctx)
	return int64(count), err
}

// GetAllowUserRegister 获取是否允许用户注册
func (r *userRepo) GetAllowUserRegister(ctx context.Context) (bool, error) {
	// 这个方法应该从系统参数读取，但为了保持接口一致性，先返回true
	// 实际实现应该在biz层调用ParamsService
	return true, nil
}

// GetByUserId 根据用户ID获取Token
func (r *userTokenRepo) GetByUserId(ctx context.Context, userId int64) (*biz.UserToken, error) {
	token, err := r.data.db.SysUserToken.Query().
		Where(sysusertoken.UserID(userId)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizToken biz.UserToken
	if err := copier.Copy(&bizToken, token); err != nil {
		return nil, err
	}
	return &bizToken, nil
}

// GetByToken 根据Token获取Token记录
func (r *userTokenRepo) GetByToken(ctx context.Context, token string) (*biz.UserToken, error) {
	tokenEntity, err := r.data.db.SysUserToken.Query().
		Where(sysusertoken.Token(token)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizToken biz.UserToken
	if err := copier.Copy(&bizToken, tokenEntity); err != nil {
		return nil, err
	}
	return &bizToken, nil
}

// Save 保存Token
func (r *userTokenRepo) Save(ctx context.Context, token *biz.UserToken) error {
	create := r.data.db.SysUserToken.Create().
		SetUserID(token.UserID).
		SetToken(token.Token).
		SetExpireDate(token.ExpireDate)

	if !token.CreateDate.IsZero() {
		create.SetCreateDate(token.CreateDate)
	}
	if !token.UpdateDate.IsZero() {
		create.SetUpdateDate(token.UpdateDate)
	}

	return create.Exec(ctx)
}

// Update 更新Token
func (r *userTokenRepo) Update(ctx context.Context, token *biz.UserToken) error {
	update := r.data.db.SysUserToken.Update().
		Where(sysusertoken.ID(token.ID))

	if token.Token != "" {
		update.SetToken(token.Token)
	}
	if !token.ExpireDate.IsZero() {
		update.SetExpireDate(token.ExpireDate)
	}
	if !token.UpdateDate.IsZero() {
		update.SetUpdateDate(token.UpdateDate)
	}

	return update.Exec(ctx)
}

// Logout 登出（使Token失效）
func (r *userTokenRepo) Logout(ctx context.Context, userId int64, expireDate time.Time) error {
	return r.data.db.SysUserToken.Update().
		Where(sysusertoken.UserID(userId)).
		SetExpireDate(expireDate).
		Exec(ctx)
}
