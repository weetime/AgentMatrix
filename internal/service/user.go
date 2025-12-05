package service

import (
	"context"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	// Version 版本号
	Version = "0.8.9"
	// DictTypeMobileArea 手机区号字典类型
	DictTypeMobileArea = "MOBILE_AREA"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	uc            *biz.UserUsecase
	configUsecase *biz.ConfigUsecase
	redisClient   *redis.Client
}

func NewUserService(
	uc *biz.UserUsecase,
	configUsecase *biz.ConfigUsecase,
	redisClientWrapper *kit.RedisClient,
) *UserService {
	return &UserService{
		uc:            uc,
		configUsecase: configUsecase,
		redisClient:   redisClientWrapper.GetClient(),
	}
}

// GetCaptcha 获取图形验证码
func (s *UserService) GetCaptcha(ctx context.Context, req *pb.GetCaptchaRequest) (*pb.Response, error) {
	uuid := req.GetUuid()
	if uuid == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "uuid is required",
		}, nil
	}

	base64Image, captchaUUID, err := kit.GenerateCaptcha(ctx, s.redisClient, uuid)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  fmt.Sprintf("failed to generate captcha: %v", err),
		}, nil
	}

	data := map[string]interface{}{
		"base64Image": base64Image,
		"uuid":        captchaUUID,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// SendSMSVerification 发送短信验证码
func (s *UserService) SendSMSVerification(ctx context.Context, req *pb.SendSMSVerificationRequest) (*pb.Response, error) {
	// 验证图形验证码
	valid := kit.ValidateCaptcha(ctx, s.redisClient, req.GetCaptchaId(), req.GetCaptcha(), false)
	if !valid {
		return &pb.Response{
			Code: 400,
			Msg:  "图形验证码错误",
		}, nil
	}

	// 检查是否开启手机注册
	isMobileRegisterStr, _ := s.configUsecase.GetValue(ctx, "server.enable_mobile_register", true)
	if isMobileRegisterStr != "true" {
		return &pb.Response{
			Code: 403,
			Msg:  "手机注册功能已关闭",
		}, nil
	}

	// 创建paramsService适配器
	paramsService := &paramsServiceAdapter{configUsecase: s.configUsecase}

	// 发送短信验证码
	err := kit.SendSMSVerificationCode(ctx, req.GetPhone(), s.redisClient, paramsService)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  fmt.Sprintf("failed to send SMS: %v", err),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.Response, error) {
	loginReq := &biz.LoginRequest{
		Username:  req.GetUsername(),
		Password:  req.GetPassword(),
		CaptchaID: req.GetCaptchaId(),
	}

	tokenDTO, err := s.uc.Login(ctx, loginReq)
	if err != nil {
		return nil, err
	}

	// 参考Java版本：Result<TokenDTO>，data直接包含token、expire、clientHash字段
	data := map[string]interface{}{
		"token":      tokenDTO.Token,
		"expire":     tokenDTO.Expire,
		"clientHash": tokenDTO.ClientHash, // 使用驼峰命名，与Java版本一致
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.Response, error) {
	// 如果开启手机注册，先验证短信验证码
	isMobileRegisterStr, _ := s.configUsecase.GetValue(ctx, "server.enable_mobile_register", true)
	if isMobileRegisterStr == "true" && req.GetMobileCaptcha() != "" {
		valid := kit.ValidateSMSVerificationCode(ctx, s.redisClient, req.GetUsername(), req.GetMobileCaptcha(), false)
		if !valid {
			return &pb.Response{
				Code: 400,
				Msg:  "短信验证码错误",
			}, nil
		}
	}

	registerReq := &biz.RegisterRequest{
		Username:      req.GetUsername(),
		Password:      req.GetPassword(),
		CaptchaID:     req.GetCaptchaId(),
		MobileCaptcha: req.GetMobileCaptcha(),
	}

	if err := s.uc.Register(ctx, registerReq); err != nil {
		return nil, err
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.Response, error) {
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	userDetail, err := s.uc.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":          userDetail.ID,
			"username":    userDetail.Username,
			"super_admin": userDetail.SuperAdmin,
			"status":      userDetail.Status,
			"token":       userDetail.Token,
		},
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.Response, error) {
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	changePasswordReq := &biz.ChangePasswordRequest{
		Password:    req.GetPassword(),
		NewPassword: req.GetNewPassword(),
	}

	if err := s.uc.ChangePassword(ctx, userID, changePasswordReq); err != nil {
		return nil, err
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// RetrievePassword 找回密码
func (s *UserService) RetrievePassword(ctx context.Context, req *pb.RetrievePasswordRequest) (*pb.Response, error) {
	// 验证短信验证码
	valid := kit.ValidateSMSVerificationCode(ctx, s.redisClient, req.GetPhone(), req.GetCode(), false)
	if !valid {
		return &pb.Response{
			Code: 400,
			Msg:  "短信验证码错误",
		}, nil
	}

	retrievePasswordReq := &biz.RetrievePasswordRequest{
		Phone:     req.GetPhone(),
		Code:      req.GetCode(),
		Password:  req.GetPassword(),
		CaptchaID: req.GetCaptchaId(),
	}

	if err := s.uc.RetrievePassword(ctx, retrievePasswordReq); err != nil {
		return nil, err
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetPubConfig 获取公共配置
func (s *UserService) GetPubConfig(ctx context.Context, req *pb.GetPubConfigRequest) (*pb.Response, error) {
	// 是否开启手机注册
	enableMobileRegisterStr, _ := s.configUsecase.GetValue(ctx, "server.enable_mobile_register", true)
	enableMobileRegister := enableMobileRegisterStr == "true"

	// 是否允许用户注册
	allowUserRegisterStr, _ := s.configUsecase.GetValue(ctx, "server.allow_user_register", true)
	allowUserRegister := allowUserRegisterStr == "true"

	// 获取手机区号列表（从字典表读取）
	mobileAreaList, _ := s.getMobileAreaList(ctx)

	// 转换为字典数据项列表（与Java版本保持一致：name和key字段）
	mobileAreaListData := make([]interface{}, 0, len(mobileAreaList))
	for _, item := range mobileAreaList {
		mobileAreaListData = append(mobileAreaListData, map[string]interface{}{
			"name": item.DictLabel, // Java版本使用name字段
			"key":  item.DictValue, // Java版本使用key字段
		})
	}

	// ICP备案号
	beianIcpNum, _ := s.configUsecase.GetValue(ctx, "server.beian_icp_num", true)

	// 公安备案号
	beianGaNum, _ := s.configUsecase.GetValue(ctx, "server.beian_ga_num", true)

	// 服务器名称
	name, _ := s.configUsecase.GetValue(ctx, "server.name", true)

	// SM2公钥
	sm2PublicKey, err := s.configUsecase.GetValue(ctx, "server.public_key", true)
	if err != nil || sm2PublicKey == "" {
		return &pb.Response{
			Code: 500,
			Msg:  "SM2公钥未配置",
		}, nil
	}

	// 当前年份
	year := fmt.Sprintf("©%d", time.Now().Year())

	// 构建响应数据
	data := map[string]interface{}{
		"enableMobileRegister": enableMobileRegister,
		"version":              Version,
		"year":                 year,
		"allowUserRegister":    allowUserRegister,
		"mobileAreaList":       mobileAreaListData,
		"beianIcpNum":          beianIcpNum,
		"beianGaNum":           beianGaNum,
		"name":                 name,
		"sm2PublicKey":         sm2PublicKey,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// getMobileAreaList 获取手机区号列表
func (s *UserService) getMobileAreaList(ctx context.Context) ([]*pb.DictDataItem, error) {
	// TODO: 实现从字典表读取手机区号列表
	// 这里先返回空列表，后续需要实现字典数据查询
	return []*pb.DictDataItem{}, nil
}

// captchaServiceAdapter 验证码服务适配器
type captchaServiceAdapter struct {
	redisClient *redis.Client
}

func (c *captchaServiceAdapter) Validate(ctx context.Context, uuid, code string, delete bool) bool {
	return kit.ValidateCaptcha(ctx, c.redisClient, uuid, code, delete)
}

// paramsServiceAdapter 参数服务适配器
type paramsServiceAdapter struct {
	configUsecase *biz.ConfigUsecase
}

func (p *paramsServiceAdapter) GetValue(paramCode string, isCache bool) (string, error) {
	return p.configUsecase.GetValue(context.Background(), paramCode, isCache)
}
