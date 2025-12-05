package service

import (
	"context"
	"fmt"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/structpb"
)

type SysParamsService struct {
	uc *biz.ConfigUsecase
	pb.UnimplementedSysParamsServiceServer
}

func NewSysParamsService(uc *biz.ConfigUsecase) *SysParamsService {
	return &SysParamsService{
		uc: uc,
	}
}

// PageSysParams 分页查询参数（使用项目规范）
func (s *SysParamsService) PageSysParams(ctx context.Context, req *pb.PageSysParamsRequest) (*pb.Response, error) {
	// 解析过滤条件

	params := &biz.ListSysParamsParams{}
	copier.Copy(params, req)
	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1 // 默认第一页
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE // 使用默认值
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))
	page.SetSortDesc() // 默认降序
	page.SetSortField(kit.DEFAULT_SORT_FIELD)

	// 查询列表
	list, err := s.uc.ListSysParams(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 查询总数
	total, err := s.uc.TotalSysParams(ctx, params)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为DTO列表
	dtoList := make([]interface{}, 0, len(list))
	for _, param := range list {
		dtoList = append(dtoList, map[string]interface{}{
			"id":         fmt.Sprintf("%d", param.ID),
			"paramCode":  param.ParamCode,
			"paramValue": param.ParamValue,
			"valueType":  param.ValueType,
			"remark":     param.Remark,
		})
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  dtoList,
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

// GetSysParams 获取参数详情
func (s *SysParamsService) GetSysParams(ctx context.Context, req *pb.GetSysParamsRequest) (*pb.Response, error) {
	param, err := s.uc.GetSysParamsByID(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 404,
			Msg:  err.Error(),
		}, nil
	}

	data := map[string]interface{}{
		"id":         fmt.Sprintf("%d", param.ID),
		"paramCode":  param.ParamCode,
		"paramValue": param.ParamValue,
		"valueType":  param.ValueType,
		"remark":     param.Remark,
	}

	dataStruct, _ := structpb.NewStruct(data)
	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// SaveSysParams 保存参数
func (s *SysParamsService) SaveSysParams(ctx context.Context, req *pb.SaveSysParamsRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	param := &biz.SysParam{}
	if err := copier.Copy(param, req); err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "数据转换失败: " + err.Error(),
		}, nil
	}
	param.ParamType = 1 // 非系统参数

	// 验证参数值
	if err := s.uc.ValidateParamValue(ctx, param.ParamCode, param.ParamValue); err != nil {
		return &pb.Response{
			Code: 400,
			Msg:  err.Error(),
		}, nil
	}

	created, err := s.uc.CreateSysParams(ctx, param)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 清除配置缓存
	s.uc.ClearConfigCache(ctx)

	data := map[string]interface{}{
		"id":         fmt.Sprintf("%d", created.ID),
		"paramCode":  created.ParamCode,
		"paramValue": created.ParamValue,
		"valueType":  created.ValueType,
		"remark":     created.Remark,
	}

	dataStruct, _ := structpb.NewStruct(data)
	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// UpdateSysParams 修改参数
func (s *SysParamsService) UpdateSysParams(ctx context.Context, req *pb.UpdateSysParamsRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	param := &biz.SysParam{}
	if err := copier.Copy(param, req); err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "数据转换失败: " + err.Error(),
		}, nil
	}

	err := s.uc.UpdateSysParams(ctx, param)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 清除配置缓存
	s.uc.ClearConfigCache(ctx)

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// DeleteSysParams 删除参数
func (s *SysParamsService) DeleteSysParams(ctx context.Context, req *pb.DeleteSysParamsRequest) (*pb.Response, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "参数ID列表不能为空",
		}, nil
	}

	err := s.uc.DeleteSysParams(ctx, ids)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 清除配置缓存
	s.uc.ClearConfigCache(ctx)

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}
