package service

import (
	"context"
	"strconv"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

type AdminService struct {
	pb.UnimplementedAdminServiceServer
	userUsecase   *biz.UserUsecase
	deviceUsecase *biz.DeviceUsecase
	userRepo      biz.UserRepo
}

func NewAdminService(
	userUsecase *biz.UserUsecase,
	deviceUsecase *biz.DeviceUsecase,
	userRepo biz.UserRepo,
) *AdminService {
	return &AdminService{
		userUsecase:   userUsecase,
		deviceUsecase: deviceUsecase,
		userRepo:      userRepo,
	}
}

// PageAdminUsers 分页查找用户
func (s *AdminService) PageAdminUsers(ctx context.Context, req *pb.PageAdminUsersRequest) (*pb.Response, error) {
	// 解析过滤条件
	var mobile *string
	if req.Mobile != nil && req.Mobile.GetValue() != "" {
		m := req.Mobile.GetValue()
		mobile = &m
	}

	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))
	page.SetSortDesc()
	page.SetSortField("id")

	// 查询列表
	list, total, err := s.userUsecase.PageAdminUsers(ctx, mobile, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := map[string]interface{}{
			"userid":      item.UserID,
			"mobile":      item.Mobile,
			"deviceCount": item.DeviceCount,
			"status":      item.Status,
			"createDate":  item.CreateDate.Format("2006-01-02 15:04:05"),
		}
		voList = append(voList, vo)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  voList,
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

// ResetUserPassword 重置用户密码
func (s *AdminService) ResetUserPassword(ctx context.Context, req *pb.ResetUserPasswordRequest) (*pb.Response, error) {
	userId, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return &pb.Response{
			Code: 400,
			Msg:  "无效的用户ID",
		}, nil
	}

	password, err := s.userUsecase.ResetPassword(ctx, userId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 返回密码字符串（直接作为data字段的值，与Java实现一致）
	// Java: Result<String>.ok(password) -> data字段直接是字符串
	// 前端代码：data.data 直接读取密码字符串
	// 在protobuf中，Response的Data是Struct类型，我们需要将字符串包装成Struct
	// 根据前端的使用方式，data.data应该直接是字符串值
	// 我们使用一个简单的key来存储密码，前端会直接读取data.data的值
	dataStruct, err := structpb.NewStruct(map[string]interface{}{
		"": password, // 使用空字符串作为key，前端会直接读取data.data
	})
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

// DeleteUser 删除用户
func (s *AdminService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.Response, error) {
	userId, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return &pb.Response{
			Code: 400,
			Msg:  "无效的用户ID",
		}, nil
	}

	err = s.userUsecase.DeleteUserById(ctx, userId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// ChangeUserStatus 批量修改用户状态
func (s *AdminService) ChangeUserStatus(ctx context.Context, req *pb.ChangeUserStatusRequest) (*pb.Response, error) {
	status := req.GetStatus()
	userIds := req.GetUserIds()

	if len(userIds) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "用户ID列表不能为空",
		}, nil
	}

	err := s.userUsecase.ChangeUserStatus(ctx, status, userIds)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// userRepoAdapter 适配器，将biz.UserRepo转换为biz.AdminUserRepo
type userRepoAdapter struct {
	repo biz.UserRepo
}

func (a *userRepoAdapter) GetUsersByIDs(ctx context.Context, userIds []int64) (map[int64]*biz.AdminUser, error) {
	users, err := a.repo.GetUsersByIDs(ctx, userIds)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*biz.AdminUser, len(users))
	for userId, user := range users {
		result[userId] = &biz.AdminUser{
			ID:       user.ID,
			Username: user.Username,
		}
	}
	return result, nil
}

// PageAdminDevices 分页查找设备
func (s *AdminService) PageAdminDevices(ctx context.Context, req *pb.PageAdminDevicesRequest) (*pb.Response, error) {
	// 解析过滤条件
	var keywords *string
	if req.Keywords != nil && req.Keywords.GetValue() != "" {
		k := req.Keywords.GetValue()
		keywords = &k
	}

	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))
	page.SetSortDesc()
	page.SetSortField("mac_address")

	// 创建适配器
	userRepoAdapter := &userRepoAdapter{repo: s.userRepo}

	// 查询列表
	list, total, err := s.deviceUsecase.PageAdminDevices(ctx, keywords, page, userRepoAdapter)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := map[string]interface{}{
			"id":             item.ID,
			"macAddress":     item.MacAddress,
			"bindUserName":   item.BindUserName,
			"deviceType":     item.DeviceType,
			"appVersion":     item.AppVersion,
			"otaUpgrade":     item.OtaUpgrade,
			"recentChatTime": item.RecentChatTime,
		}
		voList = append(voList, vo)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  voList,
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

