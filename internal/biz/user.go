package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"
	"github.com/weetime/agent-matrix/internal/middleware"

	"github.com/go-kratos/kratos/v2/log"
)

// User 用户实体
type User struct {
	ID         int64     `json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	SuperAdmin int32     `json:"super_admin"`
	Status     int32     `json:"status"`
	CreateDate time.Time `json:"create_date"`
	Updater    int64     `json:"updater"`
	Creator    int64     `json:"creator"`
	UpdateDate time.Time `json:"update_date"`
}

// UserToken Token实体
type UserToken struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Token      string    `json:"token"`
	ExpireDate time.Time `json:"expire_date"`
	UpdateDate time.Time `json:"update_date"`
	CreateDate time.Time `json:"create_date"`
}

// TokenDTO Token数据传输对象
type TokenDTO struct {
	Token      string `json:"token"`
	Expire     int32  `json:"expire"`
	ClientHash string `json:"client_hash"`
}

// UserDetail 用户详情（用于返回给前端）
type UserDetail struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	SuperAdmin int32  `json:"super_admin"`
	Status     int32  `json:"status"`
	Token      string `json:"token"`
}

// UserRepo 用户数据访问接口
type UserRepo interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByUserId(ctx context.Context, userId int64) (*User, error)
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	ChangePassword(ctx context.Context, userId int64, newPassword string) error
	ChangePasswordDirectly(ctx context.Context, userId int64, newPassword string) error
	GetUserCount(ctx context.Context) (int64, error)
	GetAllowUserRegister(ctx context.Context) (bool, error)
	GetUsersByIDs(ctx context.Context, userIds []int64) (map[int64]*User, error)
}

// UserTokenRepo Token数据访问接口
type UserTokenRepo interface {
	GetByUserId(ctx context.Context, userId int64) (*UserToken, error)
	GetByToken(ctx context.Context, token string) (*UserToken, error)
	Save(ctx context.Context, token *UserToken) error
	Update(ctx context.Context, token *UserToken) error
	Logout(ctx context.Context, userId int64, expireDate time.Time) error
}

// ParamsService 系统参数服务接口
type ParamsService interface {
	GetValue(paramCode string, isCache bool) (string, error)
}

// CaptchaService 验证码服务接口
type CaptchaService interface {
	Validate(ctx context.Context, uuid, code string, delete bool) bool
}

// SMSService 短信服务接口（暂时不需要，直接使用kit中的方法）

// UserUsecase 用户业务逻辑
type UserUsecase struct {
	userRepo       UserRepo
	tokenRepo      UserTokenRepo
	paramsService  ParamsService
	captchaService CaptchaService
	handleError    *cerrors.HandleError
	log            *log.Helper
}

// NewUserUsecase 创建UserUsecase
func NewUserUsecase(
	userRepo UserRepo,
	tokenRepo UserTokenRepo,
	paramsService ParamsService,
	captchaService CaptchaService,
	logger log.Logger,
) *UserUsecase {
	return &UserUsecase{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		paramsService:  paramsService,
		captchaService: captchaService,
		handleError:    cerrors.NewHandleError(logger),
		log:            kit.LogHelper(logger),
	}
}

// NewUserTokenService 创建UserTokenService（实现middleware.TokenService接口）
func NewUserTokenService(uc *UserUsecase) middleware.TokenService {
	return uc
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"` // SM2加密的密码
	CaptchaID string `json:"captcha_id"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"` // SM2加密的密码
	CaptchaID     string `json:"captcha_id"`
	MobileCaptcha string `json:"mobile_captcha,omitempty"` // 手机验证码（如果开启手机注册）
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

// RetrievePasswordRequest 找回密码请求
type RetrievePasswordRequest struct {
	Phone     string `json:"phone"`
	Code      string `json:"code"`     // 短信验证码
	Password  string `json:"password"` // SM2加密的新密码
	CaptchaID string `json:"captcha_id"`
}

// Login 用户登录
func (uc *UserUsecase) Login(ctx context.Context, req *LoginRequest) (*TokenDTO, error) {
	// SM2解密密码并验证验证码
	actualPassword, err := kit.DecryptAndValidateCaptcha(
		req.Password,
		req.CaptchaID,
		&captchaServiceWrapper{captchaService: uc.captchaService},
		uc.paramsService,
	)
	if err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}

	// 查询用户
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if user == nil {
		return nil, uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("账号或密码错误"))
	}

	// 验证密码
	if !kit.CheckPassword(actualPassword, user.Password) {
		return nil, uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("账号或密码错误"))
	}

	// 生成Token
	tokenDTO, err := uc.createToken(ctx, user.ID)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	return tokenDTO, nil
}

// Register 用户注册
func (uc *UserUsecase) Register(ctx context.Context, req *RegisterRequest) error {
	// 检查是否允许注册
	allowRegister, err := uc.userRepo.GetAllowUserRegister(ctx)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if !allowRegister {
		// 从系统参数读取
		allowRegisterStr, err := uc.paramsService.GetValue("server.allow_user_register", true)
		if err == nil && allowRegisterStr == "true" {
			allowRegister = true
		} else {
			return uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("用户注册功能已关闭"))
		}
	}

	// SM2解密密码并验证验证码
	actualPassword, err := kit.DecryptAndValidateCaptcha(
		req.Password,
		req.CaptchaID,
		&captchaServiceWrapper{captchaService: uc.captchaService},
		uc.paramsService,
	)
	if err != nil {
		return uc.handleError.ErrInvalidInput(ctx, err)
	}

	// 检查是否开启手机注册
	isMobileRegisterStr, _ := uc.paramsService.GetValue("server.enable_mobile_register", true)
	isMobileRegister := isMobileRegisterStr == "true"

	if isMobileRegister {
		// 验证用户是否是手机号码
		if !kit.IsValidPhone(req.Username) {
			return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("用户名必须是手机号码"))
		}

		// 短信验证码已在service层验证，这里不需要再次验证
	}

	// 检查用户是否存在
	existingUser, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if existingUser != nil {
		return uc.handleError.ErrAlreadyExists(ctx, fmt.Errorf("该手机号已注册"))
	}

	// 检查密码强度
	if !kit.IsStrongPassword(actualPassword) {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("密码强度不够，至少8位且包含字母和数字"))
	}

	// 加密密码
	hashedPassword, err := kit.HashPassword(actualPassword)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 获取用户总数，判断是否为第一个用户（超级管理员）
	userCount, err := uc.userRepo.GetUserCount(ctx)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 创建用户
	user := &User{
		Username:   req.Username,
		Password:   hashedPassword,
		SuperAdmin: 0, // 默认不是超级管理员
		Status:     1, // 默认正常状态
		CreateDate: time.Now(),
	}

	// 第一个用户设置为超级管理员
	if userCount == 0 {
		user.SuperAdmin = 1
	}

	// 保存用户
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// GetUserInfo 获取当前用户信息
func (uc *UserUsecase) GetUserInfo(ctx context.Context, userId int64) (*UserDetail, error) {
	user, err := uc.userRepo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if user == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("用户不存在"))
	}

	// 获取Token
	tokenEntity, _ := uc.tokenRepo.GetByUserId(ctx, userId)
	token := ""
	if tokenEntity != nil {
		token = tokenEntity.Token
	}

	return &UserDetail{
		ID:         user.ID,
		Username:   user.Username,
		SuperAdmin: user.SuperAdmin,
		Status:     user.Status,
		Token:      token,
	}, nil
}

// ChangePassword 修改密码
func (uc *UserUsecase) ChangePassword(ctx context.Context, userId int64, req *ChangePasswordRequest) error {
	// 获取用户
	user, err := uc.userRepo.GetByUserId(ctx, userId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if user == nil {
		return uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("用户不存在"))
	}

	// 验证旧密码
	if !kit.CheckPassword(req.Password, user.Password) {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("原密码错误"))
	}

	// 检查新密码强度
	if !kit.IsStrongPassword(req.NewPassword) {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("密码强度不够，至少8位且包含字母和数字"))
	}

	// 加密新密码
	hashedPassword, err := kit.HashPassword(req.NewPassword)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 更新密码
	if err := uc.userRepo.ChangePassword(ctx, userId, hashedPassword); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 使Token失效
	expireDate := time.Now().Add(-1 * time.Minute)
	if err := uc.tokenRepo.Logout(ctx, userId, expireDate); err != nil {
		uc.log.Warn("Failed to logout user after password change", "error", err)
	}

	return nil
}

// RetrievePassword 找回密码
func (uc *UserUsecase) RetrievePassword(ctx context.Context, req *RetrievePasswordRequest) error {
	// 检查是否开启手机注册
	isMobileRegisterStr, _ := uc.paramsService.GetValue("server.enable_mobile_register", true)
	if isMobileRegisterStr != "true" {
		return uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("找回密码功能已关闭"))
	}

	// 验证手机号格式
	if !kit.IsValidPhone(req.Phone) {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("手机号格式错误"))
	}

	// 查询用户
	user, err := uc.userRepo.GetByUsername(ctx, req.Phone)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if user == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("该手机号未注册"))
	}

	// 短信验证码已在service层验证，这里不需要再次验证

	// SM2解密新密码并验证验证码
	actualPassword, err := kit.DecryptAndValidateCaptcha(
		req.Password,
		req.CaptchaID,
		&captchaServiceWrapper{captchaService: uc.captchaService},
		uc.paramsService,
	)
	if err != nil {
		return uc.handleError.ErrInvalidInput(ctx, err)
	}

	// 检查密码强度
	if !kit.IsStrongPassword(actualPassword) {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("密码强度不够，至少8位且包含字母和数字"))
	}

	// 加密新密码
	hashedPassword, err := kit.HashPassword(actualPassword)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 直接修改密码
	if err := uc.userRepo.ChangePasswordDirectly(ctx, user.ID, hashedPassword); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// createToken 创建Token
func (uc *UserUsecase) createToken(ctx context.Context, userId int64) (*TokenDTO, error) {
	now := time.Now()
	expireTime := now.Add(kit.TokenExpireSeconds * time.Second)

	// 查询是否已有Token
	tokenEntity, err := uc.tokenRepo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	var token string
	if tokenEntity != nil && tokenEntity.ExpireDate.After(now) {
		// Token未过期，复用现有Token
		token = tokenEntity.Token
		// 更新过期时间
		tokenEntity.ExpireDate = expireTime
		tokenEntity.UpdateDate = now
		if err := uc.tokenRepo.Update(ctx, tokenEntity); err != nil {
			return nil, err
		}
	} else {
		// 生成新Token
		token = kit.GenerateToken()
		newToken := &UserToken{
			UserID:     userId,
			Token:      token,
			ExpireDate: expireTime,
			CreateDate: now,
			UpdateDate: now,
		}

		if tokenEntity != nil {
			// 更新现有Token记录
			tokenEntity.Token = token
			tokenEntity.ExpireDate = expireTime
			tokenEntity.UpdateDate = now
			if err := uc.tokenRepo.Update(ctx, tokenEntity); err != nil {
				return nil, err
			}
		} else {
			// 创建新Token记录
			if err := uc.tokenRepo.Save(ctx, newToken); err != nil {
				return nil, err
			}
		}
	}

	return &TokenDTO{
		Token:      token,
		Expire:     int32(kit.TokenExpireSeconds),
		ClientHash: "", // TODO: 从请求中获取客户端指纹
	}, nil
}

// GetUserByToken 根据Token获取用户（实现TokenService接口）
// 参考Java的Oauth2Realm.doGetAuthenticationInfo实现
func (uc *UserUsecase) GetUserByToken(ctx context.Context, token string) (*middleware.UserDetail, error) {
	// 根据accessToken，查询用户token信息
	tokenEntity, err := uc.tokenRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if tokenEntity == nil {
		return nil, fmt.Errorf("token not found")
	}

	// token失效检查（参考Java实现）
	now := time.Now()
	if tokenEntity.ExpireDate.Before(now) {
		return nil, fmt.Errorf("token expired")
	}

	// 查询用户信息
	user, err := uc.userRepo.GetByUserId(ctx, tokenEntity.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 账号状态检查（参考Java的Oauth2Realm实现）
	// status为0表示账号被锁定，会在中间件中处理
	// 这里只返回用户信息，状态检查在中间件中进行

	return &middleware.UserDetail{
		ID:         user.ID,
		Username:   user.Username,
		SuperAdmin: user.SuperAdmin,
		Status:     user.Status,
		Token:      token,
	}, nil
}

// captchaServiceWrapper 验证码服务包装器
type captchaServiceWrapper struct {
	captchaService CaptchaService
}

func (w *captchaServiceWrapper) Validate(uuid, code string, delete bool) bool {
	return w.captchaService.Validate(context.Background(), uuid, code, delete)
}
