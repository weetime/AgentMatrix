package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// SysDictType 字典类型实体
type SysDictType struct {
	ID         int64
	DictType   string
	DictName   string
	Remark     string
	Sort       int32
	Creator    int64
	CreateDate time.Time
	Updater    int64
	UpdateDate time.Time
}

// SysDictData 字典数据实体
type SysDictData struct {
	ID         int64
	DictTypeID int64
	DictLabel  string
	DictValue  string
	Remark     string
	Sort       int32
	Creator    int64
	CreateDate time.Time
	Updater    int64
	UpdateDate time.Time
}

// SysDictDataItem 字典数据项（用于根据类型获取列表）
type SysDictDataItem struct {
	Name string // dictLabel
	Key  string // dictValue
}

// ListDictTypeParams 查询字典类型过滤条件
type ListDictTypeParams struct {
	DictType *string // 字典类型编码（模糊查询）
	DictName *string // 字典名称（模糊查询）
}

// ListDictDataParams 查询字典数据过滤条件
type ListDictDataParams struct {
	DictTypeID int64   // 字典类型ID（必填）
	DictLabel  *string // 字典标签（模糊查询）
	DictValue  *string // 字典值（模糊查询）
}

// DictTypeRepo 字典类型数据访问接口
type DictTypeRepo interface {
	ListDictTypes(ctx context.Context, params *ListDictTypeParams, page *kit.PageRequest) ([]*SysDictType, error)
	TotalDictTypes(ctx context.Context, params *ListDictTypeParams) (int, error)
	GetDictTypeByID(ctx context.Context, id int64) (*SysDictType, error)
	GetDictTypeByType(ctx context.Context, dictType string) (*SysDictType, error)
	CreateDictType(ctx context.Context, dictType *SysDictType) (*SysDictType, error)
	UpdateDictType(ctx context.Context, dictType *SysDictType) error
	DeleteDictTypes(ctx context.Context, ids []int64) error
}

// DictDataRepo 字典数据数据访问接口
type DictDataRepo interface {
	ListDictData(ctx context.Context, params *ListDictDataParams, page *kit.PageRequest) ([]*SysDictData, error)
	TotalDictData(ctx context.Context, params *ListDictDataParams) (int, error)
	GetDictDataByID(ctx context.Context, id int64) (*SysDictData, error)
	GetDictDataByType(ctx context.Context, dictType string) ([]*SysDictDataItem, error)
	GetTypeByTypeID(ctx context.Context, dictTypeID int64) (string, error)
	CreateDictData(ctx context.Context, dictData *SysDictData) (*SysDictData, error)
	UpdateDictData(ctx context.Context, dictData *SysDictData) error
	DeleteDictData(ctx context.Context, ids []int64) error
	DeleteDictDataByTypeID(ctx context.Context, dictTypeID int64) error
	CheckDictValueUnique(ctx context.Context, dictTypeID int64, dictValue string, excludeID *int64) (bool, error)
}

// DictTypeUsecase 字典类型业务逻辑
type DictTypeUsecase struct {
	repo         DictTypeRepo
	dictDataRepo DictDataRepo
	userRepo     UserRepo
	handleError  *cerrors.HandleError
	log          *log.Helper
}

// NewDictTypeUsecase 创建字典类型用例
func NewDictTypeUsecase(
	repo DictTypeRepo,
	dictDataRepo DictDataRepo,
	userRepo UserRepo,
	logger log.Logger,
) *DictTypeUsecase {
	return &DictTypeUsecase{
		repo:         repo,
		dictDataRepo: dictDataRepo,
		userRepo:     userRepo,
		handleError:  cerrors.NewHandleError(logger),
		log:          kit.LogHelper(logger),
	}
}

// ListDictTypes 分页查询字典类型
func (uc *DictTypeUsecase) ListDictTypes(ctx context.Context, params *ListDictTypeParams, page *kit.PageRequest) ([]*SysDictType, error) {
	if err := kit.Validate(params); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}
	return uc.repo.ListDictTypes(ctx, params, page)
}

// TotalDictTypes 获取字典类型总数
func (uc *DictTypeUsecase) TotalDictTypes(ctx context.Context, params *ListDictTypeParams) (int, error) {
	if err := kit.Validate(params); err != nil {
		return 0, uc.handleError.ErrInvalidInput(ctx, err)
	}
	return uc.repo.TotalDictTypes(ctx, params)
}

// GetDictTypeByID 根据ID获取字典类型
func (uc *DictTypeUsecase) GetDictTypeByID(ctx context.Context, id int64) (*SysDictType, error) {
	dictType, err := uc.repo.GetDictTypeByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if dictType == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("字典类型不存在"))
	}
	return dictType, nil
}

// CreateDictType 创建字典类型
func (uc *DictTypeUsecase) CreateDictType(ctx context.Context, dictType *SysDictType) (*SysDictType, error) {
	// 检查字典类型编码是否已存在
	existing, err := uc.repo.GetDictTypeByType(ctx, dictType.DictType)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if existing != nil {
		return nil, uc.handleError.ErrAlreadyExists(ctx, fmt.Errorf("字典类型编码 %s 已存在", dictType.DictType))
	}

	// 设置默认值
	if dictType.Sort == 0 {
		dictType.Sort = 0
	}
	if dictType.CreateDate.IsZero() {
		dictType.CreateDate = time.Now()
	}
	if dictType.UpdateDate.IsZero() {
		dictType.UpdateDate = time.Now()
	}

	return uc.repo.CreateDictType(ctx, dictType)
}

// UpdateDictType 更新字典类型
func (uc *DictTypeUsecase) UpdateDictType(ctx context.Context, dictType *SysDictType) error {
	// 检查字典类型是否存在
	existing, err := uc.repo.GetDictTypeByID(ctx, dictType.ID)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if existing == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("字典类型不存在"))
	}

	// 如果修改了字典类型编码，检查新编码是否已存在
	if dictType.DictType != existing.DictType {
		typeExists, err := uc.repo.GetDictTypeByType(ctx, dictType.DictType)
		if err != nil {
			return uc.handleError.ErrInternal(ctx, err)
		}
		if typeExists != nil && typeExists.ID != dictType.ID {
			return uc.handleError.ErrAlreadyExists(ctx, fmt.Errorf("字典类型编码 %s 已存在", dictType.DictType))
		}
	}

	// 设置更新时间
	dictType.UpdateDate = time.Now()

	return uc.repo.UpdateDictType(ctx, dictType)
}

// DeleteDictTypes 批量删除字典类型
func (uc *DictTypeUsecase) DeleteDictTypes(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("字典类型ID列表不能为空"))
	}

	// 先删除对应的字典数据
	for _, id := range ids {
		if err := uc.dictDataRepo.DeleteDictDataByTypeID(ctx, id); err != nil {
			return uc.handleError.ErrInternal(ctx, err)
		}
	}

	// 再删除字典类型
	return uc.repo.DeleteDictTypes(ctx, ids)
}

// FillUserNames 填充用户名（creatorName和updaterName）
func (uc *DictTypeUsecase) FillUserNames(ctx context.Context, dictTypes []*SysDictType) error {
	if len(dictTypes) == 0 {
		return nil
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

	// 用户名填充在service层处理，这里不需要查询
	return nil
}

// DictDataUsecase 字典数据业务逻辑
type DictDataUsecase struct {
	repo         DictDataRepo
	dictTypeRepo DictTypeRepo
	userRepo     UserRepo
	handleError  *cerrors.HandleError
	log          *log.Helper
}

// NewDictDataUsecase 创建字典数据用例
func NewDictDataUsecase(
	repo DictDataRepo,
	dictTypeRepo DictTypeRepo,
	userRepo UserRepo,
	logger log.Logger,
) *DictDataUsecase {
	return &DictDataUsecase{
		repo:         repo,
		dictTypeRepo: dictTypeRepo,
		userRepo:     userRepo,
		handleError:  cerrors.NewHandleError(logger),
		log:          kit.LogHelper(logger),
	}
}

// ListDictData 分页查询字典数据
func (uc *DictDataUsecase) ListDictData(ctx context.Context, params *ListDictDataParams, page *kit.PageRequest) ([]*SysDictData, error) {
	if err := kit.Validate(params); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}
	if params.DictTypeID <= 0 {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("dictTypeId不能为空"))
	}
	return uc.repo.ListDictData(ctx, params, page)
}

// TotalDictData 获取字典数据总数
func (uc *DictDataUsecase) TotalDictData(ctx context.Context, params *ListDictDataParams) (int, error) {
	if err := kit.Validate(params); err != nil {
		return 0, uc.handleError.ErrInvalidInput(ctx, err)
	}
	if params.DictTypeID <= 0 {
		return 0, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("dictTypeId不能为空"))
	}
	return uc.repo.TotalDictData(ctx, params)
}

// GetDictDataByID 根据ID获取字典数据
func (uc *DictDataUsecase) GetDictDataByID(ctx context.Context, id int64) (*SysDictData, error) {
	dictData, err := uc.repo.GetDictDataByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if dictData == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("字典数据不存在"))
	}
	return dictData, nil
}

// GetDictDataByType 根据字典类型获取数据列表
func (uc *DictDataUsecase) GetDictDataByType(ctx context.Context, dictType string) ([]*SysDictDataItem, error) {
	if dictType == "" {
		return nil, nil
	}
	return uc.repo.GetDictDataByType(ctx, dictType)
}

// CreateDictData 创建字典数据
func (uc *DictDataUsecase) CreateDictData(ctx context.Context, dictData *SysDictData) (*SysDictData, error) {
	// 检查字典类型是否存在
	dictType, err := uc.dictTypeRepo.GetDictTypeByID(ctx, dictData.DictTypeID)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if dictType == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("字典类型不存在"))
	}

	// 检查同一字典类型下字典值是否唯一
	exists, err := uc.repo.CheckDictValueUnique(ctx, dictData.DictTypeID, dictData.DictValue, nil)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if exists {
		return nil, uc.handleError.ErrAlreadyExists(ctx, fmt.Errorf("该字典类型下字典值 %s 已存在", dictData.DictValue))
	}

	// 设置默认值
	if dictData.Sort == 0 {
		dictData.Sort = 0
	}
	if dictData.CreateDate.IsZero() {
		dictData.CreateDate = time.Now()
	}
	if dictData.UpdateDate.IsZero() {
		dictData.UpdateDate = time.Now()
	}

	return uc.repo.CreateDictData(ctx, dictData)
}

// UpdateDictData 更新字典数据
func (uc *DictDataUsecase) UpdateDictData(ctx context.Context, dictData *SysDictData) error {
	// 检查字典数据是否存在
	existing, err := uc.repo.GetDictDataByID(ctx, dictData.ID)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if existing == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("字典数据不存在"))
	}

	// 检查字典类型是否存在
	dictType, err := uc.dictTypeRepo.GetDictTypeByID(ctx, dictData.DictTypeID)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if dictType == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("字典类型不存在"))
	}

	// 如果修改了字典值，检查新值是否唯一
	if dictData.DictValue != existing.DictValue {
		excludeID := &dictData.ID
		exists, err := uc.repo.CheckDictValueUnique(ctx, dictData.DictTypeID, dictData.DictValue, excludeID)
		if err != nil {
			return uc.handleError.ErrInternal(ctx, err)
		}
		if exists {
			return uc.handleError.ErrAlreadyExists(ctx, fmt.Errorf("该字典类型下字典值 %s 已存在", dictData.DictValue))
		}
	}

	// 设置更新时间
	dictData.UpdateDate = time.Now()

	return uc.repo.UpdateDictData(ctx, dictData)
}

// DeleteDictData 批量删除字典数据
func (uc *DictDataUsecase) DeleteDictData(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("字典数据ID列表不能为空"))
	}
	return uc.repo.DeleteDictData(ctx, ids)
}

// GetTypeByTypeID 根据字典类型ID获取字典类型编码
func (uc *DictDataUsecase) GetTypeByTypeID(ctx context.Context, dictTypeID int64) (string, error) {
	return uc.repo.GetTypeByTypeID(ctx, dictTypeID)
}
