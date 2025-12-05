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

type SysDictTypeService struct {
	uc          *biz.DictTypeUsecase
	userRepo    biz.UserRepo
	redisClient *redis.Client
	pb.UnimplementedSysDictTypeServiceServer
}

type SysDictDataService struct {
	uc          *biz.DictDataUsecase
	userRepo    biz.UserRepo
	redisClient *redis.Client
	pb.UnimplementedSysDictDataServiceServer
}

func NewSysDictTypeService(
	uc *biz.DictTypeUsecase,
	userRepo biz.UserRepo,
	redisClientWrapper *kit.RedisClient,
) *SysDictTypeService {
	return &SysDictTypeService{
		uc:          uc,
		userRepo:    userRepo,
		redisClient: redisClientWrapper.GetClient(),
	}
}

func NewSysDictDataService(
	uc *biz.DictDataUsecase,
	userRepo biz.UserRepo,
	redisClientWrapper *kit.RedisClient,
) *SysDictDataService {
	return &SysDictDataService{
		uc:          uc,
		userRepo:    userRepo,
		redisClient: redisClientWrapper.GetClient(),
	}
}

// PageSysDictType 分页查询字典类型
func (s *SysDictTypeService) PageSysDictType(ctx context.Context, req *pb.PageSysDictTypeRequest) (*pb.Response, error) {
	// 解析过滤条件
	params := &biz.ListDictTypeParams{}
	if req.DictType != nil && req.DictType.GetValue() != "" {
		dictType := req.DictType.GetValue()
		params.DictType = &dictType
	}
	if req.DictName != nil && req.DictName.GetValue() != "" {
		dictName := req.DictName.GetValue()
		params.DictName = &dictName
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
	list, err := s.uc.ListDictTypes(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 查询总数
	total, err := s.uc.TotalDictTypes(ctx, params)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 批量获取用户名映射
	userNameMap := s.getUserNameMap(ctx, list)

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := s.dictTypeToVOWithUserNames(item, userNameMap)
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

// GetSysDictType 获取字典类型详情
func (s *SysDictTypeService) GetSysDictType(ctx context.Context, req *pb.GetSysDictTypeRequest) (*pb.Response, error) {
	dictType, err := s.uc.GetDictTypeByID(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 404,
			Msg:  err.Error(),
		}, nil
	}

	// 获取用户名映射
	userNameMap := s.getUserNameMap(ctx, []*biz.SysDictType{dictType})

	data := s.dictTypeToVOWithUserNames(dictType, userNameMap)
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

// SaveSysDictType 保存字典类型
func (s *SysDictTypeService) SaveSysDictType(ctx context.Context, req *pb.SaveSysDictTypeRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	dictType := &biz.SysDictType{}
	dictType.DictType = req.GetDictType()
	dictType.DictName = req.GetDictName()
	dictType.Remark = req.GetRemark()
	dictType.Sort = req.GetSort()

	// 从context获取用户ID
	if userID := getUserIDFromContext(ctx); userID > 0 {
		dictType.Creator = userID
		dictType.Updater = userID
	}

	created, err := s.uc.CreateDictType(ctx, dictType)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 获取用户名映射
	userNameMap := s.getUserNameMap(ctx, []*biz.SysDictType{created})
	data := s.dictTypeToVOWithUserNames(created, userNameMap)
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

// UpdateSysDictType 修改字典类型
func (s *SysDictTypeService) UpdateSysDictType(ctx context.Context, req *pb.UpdateSysDictTypeRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	dictType := &biz.SysDictType{}
	dictType.ID = req.GetId()
	dictType.DictType = req.GetDictType()
	dictType.DictName = req.GetDictName()
	dictType.Remark = req.GetRemark()
	dictType.Sort = req.GetSort()

	// 从context获取用户ID
	if userID := getUserIDFromContext(ctx); userID > 0 {
		dictType.Updater = userID
	}

	err := s.uc.UpdateDictType(ctx, dictType)
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

// DeleteSysDictType 删除字典类型
func (s *SysDictTypeService) DeleteSysDictType(ctx context.Context, req *pb.DeleteSysDictTypeRequest) (*pb.Response, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "字典类型ID列表不能为空",
		}, nil
	}

	err := s.uc.DeleteDictTypes(ctx, ids)
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

// PageSysDictData 分页查询字典数据
func (s *SysDictDataService) PageSysDictData(ctx context.Context, req *pb.PageSysDictDataRequest) (*pb.Response, error) {
	// 解析过滤条件
	params := &biz.ListDictDataParams{
		DictTypeID: req.GetDictTypeId(),
	}
	if req.DictLabel != nil && req.DictLabel.GetValue() != "" {
		dictLabel := req.DictLabel.GetValue()
		params.DictLabel = &dictLabel
	}
	if req.DictValue != nil && req.DictValue.GetValue() != "" {
		dictValue := req.DictValue.GetValue()
		params.DictValue = &dictValue
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
	list, err := s.uc.ListDictData(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 查询总数
	total, err := s.uc.TotalDictData(ctx, params)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 批量获取用户名映射
	userNameMap := s.getUserNameMap(ctx, list)

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := s.dictDataToVOWithUserNames(item, userNameMap)
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

// GetSysDictData 获取字典数据详情
func (s *SysDictDataService) GetSysDictData(ctx context.Context, req *pb.GetSysDictDataRequest) (*pb.Response, error) {
	dictData, err := s.uc.GetDictDataByID(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 404,
			Msg:  err.Error(),
		}, nil
	}

	// 获取用户名映射
	userNameMap := s.getUserNameMap(ctx, []*biz.SysDictData{dictData})

	data := s.dictDataToVOWithUserNames(dictData, userNameMap)
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

// SaveSysDictData 新增字典数据
func (s *SysDictDataService) SaveSysDictData(ctx context.Context, req *pb.SaveSysDictDataRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	dictData := &biz.SysDictData{}
	dictData.DictTypeID = req.GetDictTypeId()
	dictData.DictLabel = req.GetDictLabel()
	dictData.DictValue = req.GetDictValue()
	dictData.Remark = req.GetRemark()
	dictData.Sort = req.GetSort()

	// 从context获取用户ID
	if userID := getUserIDFromContext(ctx); userID > 0 {
		dictData.Creator = userID
		dictData.Updater = userID
	}

	created, err := s.uc.CreateDictData(ctx, dictData)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 清除Redis缓存
	if err := s.clearDictDataCache(ctx, created.DictTypeID); err != nil {
		// 记录日志但不影响主流程
		fmt.Printf("清除字典数据缓存失败: %v\n", err)
	}

	// 获取用户名映射
	userNameMap := s.getUserNameMap(ctx, []*biz.SysDictData{created})
	data := s.dictDataToVOWithUserNames(created, userNameMap)
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

// UpdateSysDictData 修改字典数据
func (s *SysDictDataService) UpdateSysDictData(ctx context.Context, req *pb.UpdateSysDictDataRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	dictData := &biz.SysDictData{}
	dictData.ID = req.GetId()
	dictData.DictTypeID = req.GetDictTypeId()
	dictData.DictLabel = req.GetDictLabel()
	dictData.DictValue = req.GetDictValue()
	dictData.Remark = req.GetRemark()
	dictData.Sort = req.GetSort()

	// 从context获取用户ID
	if userID := getUserIDFromContext(ctx); userID > 0 {
		dictData.Updater = userID
	}

	err := s.uc.UpdateDictData(ctx, dictData)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 清除Redis缓存
	if err := s.clearDictDataCache(ctx, dictData.DictTypeID); err != nil {
		// 记录日志但不影响主流程
		fmt.Printf("清除字典数据缓存失败: %v\n", err)
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// DeleteSysDictData 删除字典数据
func (s *SysDictDataService) DeleteSysDictData(ctx context.Context, req *pb.DeleteSysDictDataRequest) (*pb.Response, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "字典数据ID列表不能为空",
		}, nil
	}

	// 先查询要删除的数据，获取dictTypeID用于清除缓存
	dictDataList := make([]*biz.SysDictData, 0, len(ids))
	for _, id := range ids {
		dictData, err := s.uc.GetDictDataByID(ctx, id)
		if err == nil && dictData != nil {
			dictDataList = append(dictDataList, dictData)
		}
	}

	err := s.uc.DeleteDictData(ctx, ids)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 清除Redis缓存
	for _, dictData := range dictDataList {
		if err := s.clearDictDataCache(ctx, dictData.DictTypeID); err != nil {
			// 记录日志但不影响主流程
			fmt.Printf("清除字典数据缓存失败: %v\n", err)
		}
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetDictDataByType 根据字典类型获取数据列表
func (s *SysDictDataService) GetDictDataByType(ctx context.Context, req *pb.GetDictDataByTypeRequest) (*pb.Response, error) {
	dictType := req.GetDictType()
	if dictType == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "字典类型不能为空",
		}, nil
	}

	// 先从Redis获取缓存
	cacheKey := kit.GetDictDataByTypeKey(dictType)
	var cachedData []*biz.SysDictDataItem
	err := kit.GetRedisObject(ctx, s.redisClient, cacheKey, &cachedData)
	if err == nil && cachedData != nil {
		// 转换为响应格式
		itemList := make([]interface{}, 0, len(cachedData))
		for _, item := range cachedData {
			itemList = append(itemList, map[string]interface{}{
				"name": item.Name,
				"key":  item.Key,
			})
		}

		data := map[string]interface{}{
			"list": itemList,
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

	// 如果缓存中没有，则从数据库获取
	data, err := s.uc.GetDictDataByType(ctx, dictType)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 存入Redis缓存
	if data != nil {
		if err := kit.SetRedisObject(ctx, s.redisClient, cacheKey, data, 0); err != nil {
			// 记录日志但不影响主流程
			fmt.Printf("保存字典数据缓存失败: %v\n", err)
		}
	}

	// 转换为响应格式
	itemList := make([]interface{}, 0)
	if data != nil {
		for _, item := range data {
			itemList = append(itemList, map[string]interface{}{
				"name": item.Name,
				"key":  item.Key,
			})
		}
	}

	responseData := map[string]interface{}{
		"list": itemList,
	}

	dataStruct, err := structpb.NewStruct(responseData)
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

// getUserNameMap 批量获取用户名映射（字典类型）
func (s *SysDictTypeService) getUserNameMap(ctx context.Context, dictTypes []*biz.SysDictType) map[int64]string {
	if len(dictTypes) == 0 {
		return make(map[int64]string)
	}

	// 收集所有用户ID
	userIDs := make([]int64, 0)
	userIDSet := make(map[int64]bool)
	for _, dt := range dictTypes {
		if dt.Creator > 0 && !userIDSet[dt.Creator] {
			userIDs = append(userIDs, dt.Creator)
			userIDSet[dt.Creator] = true
		}
		if dt.Updater > 0 && !userIDSet[dt.Updater] {
			userIDs = append(userIDs, dt.Updater)
			userIDSet[dt.Updater] = true
		}
	}

	if len(userIDs) == 0 {
		return make(map[int64]string)
	}

	// 批量查询用户信息
	userMap, err := s.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return make(map[int64]string)
	}

	// 转换为用户名映射
	userNameMap := make(map[int64]string)
	for id, user := range userMap {
		if user != nil {
			userNameMap[id] = user.Username
		}
	}

	return userNameMap
}

// getUserNameMap 批量获取用户名映射（字典数据）
func (s *SysDictDataService) getUserNameMap(ctx context.Context, dictDataList []*biz.SysDictData) map[int64]string {
	if len(dictDataList) == 0 {
		return make(map[int64]string)
	}

	// 收集所有用户ID
	userIDs := make([]int64, 0)
	userIDSet := make(map[int64]bool)
	for _, dd := range dictDataList {
		if dd.Creator > 0 && !userIDSet[dd.Creator] {
			userIDs = append(userIDs, dd.Creator)
			userIDSet[dd.Creator] = true
		}
		if dd.Updater > 0 && !userIDSet[dd.Updater] {
			userIDs = append(userIDs, dd.Updater)
			userIDSet[dd.Updater] = true
		}
	}

	if len(userIDs) == 0 {
		return make(map[int64]string)
	}

	// 批量查询用户信息
	userMap, err := s.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return make(map[int64]string)
	}

	// 转换为用户名映射
	userNameMap := make(map[int64]string)
	for id, user := range userMap {
		if user != nil {
			userNameMap[id] = user.Username
		}
	}

	return userNameMap
}

// dictTypeToVOWithUserNames 转换为VO格式（带用户名）
func (s *SysDictTypeService) dictTypeToVOWithUserNames(dictType *biz.SysDictType, userNameMap map[int64]string) map[string]interface{} {
	creatorName := userNameMap[dictType.Creator]
	updaterName := userNameMap[dictType.Updater]

	return map[string]interface{}{
		"id":          fmt.Sprintf("%d", dictType.ID),
		"dictType":    dictType.DictType,
		"dictName":    dictType.DictName,
		"remark":      dictType.Remark,
		"sort":        dictType.Sort,
		"creator":     fmt.Sprintf("%d", dictType.Creator),
		"creatorName": creatorName,
		"createDate":  formatTime(dictType.CreateDate),
		"updater":     fmt.Sprintf("%d", dictType.Updater),
		"updaterName": updaterName,
		"updateDate":  formatTime(dictType.UpdateDate),
	}
}

// dictDataToVOWithUserNames 转换为VO格式（带用户名）
func (s *SysDictDataService) dictDataToVOWithUserNames(dictData *biz.SysDictData, userNameMap map[int64]string) map[string]interface{} {
	creatorName := userNameMap[dictData.Creator]
	updaterName := userNameMap[dictData.Updater]

	return map[string]interface{}{
		"id":          fmt.Sprintf("%d", dictData.ID),
		"dictTypeId":  fmt.Sprintf("%d", dictData.DictTypeID),
		"dictLabel":   dictData.DictLabel,
		"dictValue":   dictData.DictValue,
		"remark":      dictData.Remark,
		"sort":        dictData.Sort,
		"creator":     fmt.Sprintf("%d", dictData.Creator),
		"creatorName": creatorName,
		"createDate":  formatTime(dictData.CreateDate),
		"updater":     fmt.Sprintf("%d", dictData.Updater),
		"updaterName": updaterName,
		"updateDate":  formatTime(dictData.UpdateDate),
	}
}

// clearDictDataCache 清除字典数据缓存
func (s *SysDictDataService) clearDictDataCache(ctx context.Context, dictTypeID int64) error {
	// 获取字典类型编码
	dictType, err := s.uc.GetTypeByTypeID(ctx, dictTypeID)
	if err != nil {
		return err
	}

	// 删除Redis缓存
	cacheKey := kit.GetDictDataByTypeKey(dictType)
	return s.redisClient.Del(ctx, cacheKey).Err()
}

// getUserIDFromContext 从context获取用户ID
func getUserIDFromContext(ctx context.Context) int64 {
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return 0
	}
	return userID
}

// formatTime 格式化时间为字符串
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
