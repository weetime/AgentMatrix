package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type VoiceCloneService struct {
	uc          *biz.VoiceCloneUsecase
	redisClient *kit.RedisClient
	pb.UnimplementedVoiceCloneServiceServer
}

func NewVoiceCloneService(
	uc *biz.VoiceCloneUsecase,
	redisClient *kit.RedisClient,
) *VoiceCloneService {
	return &VoiceCloneService{
		uc:          uc,
		redisClient: redisClient,
	}
}

// PageVoiceClone 分页查询音色资源
func (s *VoiceCloneService) PageVoiceClone(ctx context.Context, req *pb.PageVoiceCloneRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 解析搜索条件
	params := &biz.ListVoiceCloneParams{
		UserID: currentUserId,
	}
	if req.Name != nil && req.Name.GetValue() != "" {
		name := req.Name.GetValue()
		params.Name = &name
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
	page.SetSortField("create_date")

	// 查询列表
	list, total, err := s.uc.PageVoiceClone(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := s.voiceCloneToVO(item)
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

// UploadVoice 上传音频进行声音克隆
// 注意：文件上传需要通过HTTP handler处理multipart/form-data
func (s *VoiceCloneService) UploadVoice(ctx context.Context, req *emptypb.Empty) (*pb.Response, error) {
	// 这个方法不应该被直接调用，文件上传需要通过HTTP handler处理
	return &pb.Response{
		Code: 400,
		Msg:  "文件上传需要通过HTTP multipart/form-data处理，请使用HTTP handler",
	}, nil
}

// UpdateVoiceCloneName 更新声音克隆名称
func (s *VoiceCloneService) UpdateVoiceCloneName(ctx context.Context, req *pb.UpdateVoiceCloneNameRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	id := req.GetId()
	name := req.GetName()

	if id == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "id不能为空",
		}, nil
	}
	if name == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "名称不能为空",
		}, nil
	}

	// 检查权限
	if err := s.checkPermission(ctx, id); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 更新名称
	err := s.uc.UpdateName(ctx, id, name)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 清除Redis缓存
	redisKey := kit.GetTimbreNameByIdKey(id)
	_ = s.redisClient.Delete(ctx, redisKey)

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetAudioId 获取音频下载ID
func (s *VoiceCloneService) GetAudioId(ctx context.Context, req *pb.GetAudioIdRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	id := req.GetId()
	if id == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "id不能为空",
		}, nil
	}

	// 检查权限
	if err := s.checkPermission(ctx, id); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 验证音频数据存在
	voiceData, err := s.uc.GetVoiceData(ctx, id)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}
	if voiceData == nil || len(voiceData) == 0 {
		return &pb.Response{
			Code: 404,
			Msg:  "音频数据不存在",
		}, nil
	}

	// 生成UUID
	uuidStr := uuid.New().String()

	// 存储到Redis，过期时间24小时
	redisKey := kit.GetVoiceCloneAudioIdKey(uuidStr)
	err = s.redisClient.Set(ctx, redisKey, id, 24*time.Hour)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "存储音频ID失败: " + err.Error(),
		}, nil
	}

	// 返回UUID
	dataStruct, err := structpb.NewStruct(map[string]interface{}{
		"uuid": uuidStr,
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

// PlayVoice 播放音频（公开接口）
func (s *VoiceCloneService) PlayVoice(ctx context.Context, req *pb.PlayVoiceRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	uuidStr := req.GetUuid()
	if uuidStr == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "uuid不能为空",
		}, nil
	}

	// 从Redis获取音频ID
	redisKey := kit.GetVoiceCloneAudioIdKey(uuidStr)
	id, err := s.redisClient.Get(ctx, redisKey)
	if err != nil || id == "" {
		return &pb.Response{
			Code: 404,
			Msg:  "音频ID不存在或已过期",
		}, nil
	}

	// 删除Redis key（一次性使用）
	_ = s.redisClient.Delete(ctx, redisKey)

	// 获取音频数据
	voiceData, err := s.uc.GetVoiceData(ctx, id)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}
	if voiceData == nil || len(voiceData) == 0 {
		return &pb.Response{
			Code: 404,
			Msg:  "音频数据不存在",
		}, nil
	}

	// 注意：音频流返回需要通过HTTP handler处理
	// 这里返回错误，提示需要使用HTTP handler
	return &pb.Response{
		Code: 400,
		Msg:  "音频播放需要通过HTTP handler处理，请使用HTTP handler",
	}, nil
}

// CloneAudio 复刻音频
func (s *VoiceCloneService) CloneAudio(ctx context.Context, req *pb.CloneAudioRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	cloneId := req.GetCloneId()
	if cloneId == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "cloneId不能为空",
		}, nil
	}

	// 检查权限
	if err := s.checkPermission(ctx, cloneId); err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 调用biz层进行语音克隆训练
	err := s.uc.CloneAudio(ctx, cloneId)
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

// voiceCloneToVO 将VoiceCloneResponseDTO转换为VO，确保字段与Java的VoiceCloneResponseDTO一致
func (s *VoiceCloneService) voiceCloneToVO(dto *biz.VoiceCloneResponseDTO) map[string]interface{} {
	return map[string]interface{}{
		"id":          dto.ID,
		"name":        dto.Name,
		"modelId":     dto.ModelID,
		"modelName":   dto.ModelName,
		"voiceId":     dto.VoiceID,
		"userId":      dto.UserID, // 已经是字符串格式
		"userName":    dto.UserName,
		"trainStatus": dto.TrainStatus,
		"trainError":  dto.TrainError,
		"createDate":  dto.CreateDate,
		"hasVoice":    dto.HasVoice,
	}
}

// checkPermission 检查权限（验证记录存在且属于当前用户）
func (s *VoiceCloneService) checkPermission(ctx context.Context, id string) error {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return fmt.Errorf("未授权")
	}

	// 查询声音克隆记录
	entity, err := s.uc.GetVoiceCloneByID(ctx, id)
	if err != nil {
		return fmt.Errorf("声音克隆记录不存在")
	}
	if entity == nil {
		return fmt.Errorf("声音克隆记录不存在")
	}

	// 检查权限：用户只能操作自己的记录
	userId, err := parseUserID(entity.UserID)
	if err != nil {
		return fmt.Errorf("用户ID格式错误")
	}
	if userId != currentUserId {
		return fmt.Errorf("无权限")
	}

	return nil
}

// parseUserID 解析用户ID字符串为int64
func parseUserID(userIDStr string) (int64, error) {
	var userId int64
	_, err := fmt.Sscanf(userIDStr, "%d", &userId)
	return userId, err
}
