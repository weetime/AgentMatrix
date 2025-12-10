package service

import (
	"context"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type OtaService struct {
	uc           *biz.OtaUsecase
	redisClient  *kit.RedisClient
	tokenService middleware.TokenService
	pb.UnimplementedOtaServiceServer
}

func NewOtaService(
	uc *biz.OtaUsecase,
	redisClient *kit.RedisClient,
	tokenService middleware.TokenService,
) *OtaService {
	return &OtaService{
		uc:           uc,
		redisClient:  redisClient,
		tokenService: tokenService,
	}
}

// checkSuperAdminPermission 检查是否为超级管理员
func (s *OtaService) checkSuperAdminPermission(ctx context.Context) error {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return fmt.Errorf("未授权")
	}
	if user.SuperAdmin != 1 {
		return fmt.Errorf("需要超级管理员权限")
	}
	return nil
}

// PageOta 分页查询OTA固件
func (s *OtaService) PageOta(ctx context.Context, req *pb.PageOtaRequest) (*pb.Response, error) {
	// 权限检查：超级管理员
	if err := s.checkSuperAdminPermission(ctx); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 解析搜索条件
	params := &biz.ListOtaParams{}
	if req.FirmwareName != nil && req.FirmwareName.GetValue() != "" {
		firmwareName := req.FirmwareName.GetValue()
		params.FirmwareName = &firmwareName
	}

	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1 // 默认第1页
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))
	page.SetSortDesc()
	page.SetSortField("update_date")

	// 查询列表
	list, total, err := s.uc.PageOta(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := s.otaToVO(item)
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

// GetOta 获取OTA固件详情
func (s *OtaService) GetOta(ctx context.Context, req *pb.GetOtaRequest) (*pb.Response, error) {
	// 权限检查：超级管理员
	if err := s.checkSuperAdminPermission(ctx); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	id := req.GetId()
	if id == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "id不能为空",
		}, nil
	}

	// 查询详情
	dto, err := s.uc.GetOtaByID(ctx, id)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}
	if dto == nil {
		return &pb.Response{
			Code: 404,
			Msg:  "OTA固件不存在",
		}, nil
	}

	// 转换为VO
	vo := s.otaToVO(dto)

	// 构建响应数据
	dataStruct, err := structpb.NewStruct(vo)
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

// SaveOta 保存OTA固件信息
func (s *OtaService) SaveOta(ctx context.Context, req *pb.SaveOtaRequest) (*pb.Response, error) {
	// 权限检查：超级管理员
	if err := s.checkSuperAdminPermission(ctx); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 获取当前用户ID
	currentUserId, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 构建实体
	entity := &biz.Ota{
		FirmwareName: req.GetFirmwareName(),
		Type:         req.GetType(),
		Version:      req.GetVersion(),
		Size:         req.GetSize(),
		Remark:       req.GetRemark(),
		FirmwarePath: req.GetFirmwarePath(),
		Sort:         req.GetSort(),
		Creator:      currentUserId,
		Updater:      currentUserId,
		CreateDate:   time.Now(),
		UpdateDate:   time.Now(),
	}

	// 调用biz层保存
	if err := s.uc.SaveOta(ctx, entity); err != nil {
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

// UpdateOta 修改OTA固件信息
func (s *OtaService) UpdateOta(ctx context.Context, req *pb.UpdateOtaRequest) (*pb.Response, error) {
	// 权限检查：超级管理员
	if err := s.checkSuperAdminPermission(ctx); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 获取当前用户ID
	currentUserId, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	id := req.GetId()
	if id == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "id不能为空",
		}, nil
	}

	// 构建实体
	entity := &biz.Ota{
		FirmwareName: req.GetFirmwareName(),
		Type:         req.GetType(),
		Version:      req.GetVersion(),
		Size:         req.GetSize(),
		Remark:       req.GetRemark(),
		FirmwarePath: req.GetFirmwarePath(),
		Sort:         req.GetSort(),
		Updater:      currentUserId,
		UpdateDate:   time.Now(),
	}

	// 调用biz层更新
	if err := s.uc.UpdateOta(ctx, id, entity); err != nil {
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

// DeleteOta 删除OTA固件
func (s *OtaService) DeleteOta(ctx context.Context, req *pb.DeleteOtaRequest) (*pb.Response, error) {
	// 权限检查：超级管理员
	if err := s.checkSuperAdminPermission(ctx); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	id := req.GetId()
	if id == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "删除的固件ID不能为空",
		}, nil
	}

	// 调用biz层删除
	if err := s.uc.DeleteOta(ctx, []string{id}); err != nil {
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

// GetDownloadUrl 获取下载链接
func (s *OtaService) GetDownloadUrl(ctx context.Context, req *pb.GetDownloadUrlRequest) (*pb.Response, error) {
	// 权限检查：超级管理员
	if err := s.checkSuperAdminPermission(ctx); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	id := req.GetId()
	if id == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "id不能为空",
		}, nil
	}

	// 调用biz层生成下载链接
	uuidStr, err := s.uc.GetDownloadUrl(ctx, id)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建响应数据
	data := map[string]interface{}{
		"uuid": uuidStr,
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

// DownloadOta 下载固件（通过HTTP handler实现）
func (s *OtaService) DownloadOta(ctx context.Context, req *pb.DownloadOtaRequest) (*pb.Response, error) {
	// 这个方法不应该被直接调用，文件下载需要通过HTTP handler处理
	return &pb.Response{
		Code: 400,
		Msg:  "文件下载需要通过HTTP handler处理",
	}, nil
}

// UploadFirmware 上传固件（通过HTTP handler实现）
func (s *OtaService) UploadFirmware(ctx context.Context, req *emptypb.Empty) (*pb.Response, error) {
	// 这个方法不应该被直接调用，文件上传需要通过HTTP handler处理
	return &pb.Response{
		Code: 400,
		Msg:  "文件上传需要通过HTTP multipart/form-data处理，请使用HTTP handler",
	}, nil
}

// otaToVO 转换为VO（确保所有ID字段格式化为字符串）
func (s *OtaService) otaToVO(dto *biz.OtaResponseDTO) map[string]interface{} {
	vo := map[string]interface{}{
		"id":           dto.ID,
		"firmwareName": dto.FirmwareName,
		"type":         dto.Type,
		"version":      dto.Version,
		"size":         dto.Size,
		"remark":       dto.Remark,
		"firmwarePath": dto.FirmwarePath,
		"sort":         dto.Sort,
		"createDate":   dto.CreateDate,
		"updateDate":   dto.UpdateDate,
	}

	// ID字段已经格式化为字符串（在DTO中处理）
	if dto.Creator != "" {
		vo["creator"] = dto.Creator
	}
	if dto.Updater != "" {
		vo["updater"] = dto.Updater
	}

	return vo
}
