package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

type TtsVoiceService struct {
	uc *biz.TtsVoiceUsecase
	pb.UnimplementedTtsVoiceServiceServer
}

func NewTtsVoiceService(
	uc *biz.TtsVoiceUsecase,
) *TtsVoiceService {
	return &TtsVoiceService{
		uc: uc,
	}
}

// PageTtsVoice 分页查找音色
func (s *TtsVoiceService) PageTtsVoice(ctx context.Context, req *pb.PageTtsVoiceRequest) (*pb.Response, error) {
	// 解析过滤条件
	params := &biz.ListTtsVoiceParams{
		TtsModelID: req.GetTtsModelId(),
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
	page.SetSortAsc()
	page.SetSortField("sort")

	// 查询列表
	list, err := s.uc.ListTtsVoice(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 查询总数
	total, err := s.uc.TotalTtsVoice(ctx, params)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := s.ttsVoiceToVO(item)
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

// SaveTtsVoice 保存音色
func (s *TtsVoiceService) SaveTtsVoice(ctx context.Context, req *pb.SaveTtsVoiceRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	// 从context获取用户ID
	userID, _ := middleware.GetUserIdFromContext(ctx)

	// 生成ID（UUID去掉横线，32位）
	id := strings.ReplaceAll(uuid.New().String(), "-", "")

	voice := &biz.TtsVoice{
		ID:             id,
		TtsModelID:     req.GetTtsModelId(),
		Name:           req.GetName(),
		TtsVoice:       req.GetTtsVoice(),
		Languages:      req.GetLanguages(),
		VoiceDemo:      req.GetVoiceDemo(),
		Remark:         req.GetRemark(),
		ReferenceAudio: req.GetReferenceAudio(),
		ReferenceText:  req.GetReferenceText(),
		Sort:           req.GetSort(),
		Creator:        userID,
		CreateDate:     time.Now(),
		UpdateDate:     time.Now(),
	}

	_, err := s.uc.CreateTtsVoice(ctx, voice)
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

// UpdateTtsVoice 修改音色
func (s *TtsVoiceService) UpdateTtsVoice(ctx context.Context, req *pb.UpdateTtsVoiceRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	// 从context获取用户ID
	userID, _ := middleware.GetUserIdFromContext(ctx)

	voice := &biz.TtsVoice{
		ID:             req.GetId(),
		TtsModelID:     req.GetTtsModelId(),
		Name:           req.GetName(),
		TtsVoice:       req.GetTtsVoice(),
		Languages:      req.GetLanguages(),
		VoiceDemo:      req.GetVoiceDemo(),
		Remark:         req.GetRemark(),
		ReferenceAudio: req.GetReferenceAudio(),
		ReferenceText:  req.GetReferenceText(),
		Sort:           req.GetSort(),
		Updater:        userID,
		UpdateDate:     time.Now(),
	}

	err := s.uc.UpdateTtsVoice(ctx, voice)
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

// DeleteTtsVoice 删除音色
func (s *TtsVoiceService) DeleteTtsVoice(ctx context.Context, req *pb.DeleteTtsVoiceRequest) (*pb.Response, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "音色ID列表不能为空",
		}, nil
	}

	err := s.uc.DeleteTtsVoice(ctx, ids)
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

// ttsVoiceToVO 将biz实体转换为VO
func (s *TtsVoiceService) ttsVoiceToVO(voice *biz.TtsVoice) map[string]interface{} {
	createDate := ""
	if !voice.CreateDate.IsZero() {
		createDate = voice.CreateDate.Format("2006-01-02 15:04:05")
	}
	updateDate := ""
	if !voice.UpdateDate.IsZero() {
		updateDate = voice.UpdateDate.Format("2006-01-02 15:04:05")
	}

	return map[string]interface{}{
		"id":             voice.ID, // ID已经是String类型
		"languages":      voice.Languages,
		"name":           voice.Name,
		"remark":         voice.Remark,
		"referenceAudio": voice.ReferenceAudio,
		"referenceText":  voice.ReferenceText,
		"sort":           voice.Sort,
		"ttsModelId":     voice.TtsModelID,
		"ttsVoice":       voice.TtsVoice,
		"voiceDemo":      voice.VoiceDemo,
		"creator":        fmt.Sprintf("%d", voice.Creator),
		"createDate":     createDate,
		"updater":        fmt.Sprintf("%d", voice.Updater),
		"updateDate":     updateDate,
	}
}
