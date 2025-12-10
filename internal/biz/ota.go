package biz

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

var (
	ErrDuplicateOtaTypeVersion = fmt.Errorf("已存在相同类型和版本的固件，请修改后重试")
)

// Ota OTA固件实体
type Ota struct {
	ID           string    // ID
	FirmwareName string    // 固件名称
	Type         string    // 固件类型
	Version      string    // 版本号
	Size         int64     // 文件大小(字节)
	Remark       string    // 备注/说明
	FirmwarePath string    // 固件路径
	Sort         int32     // 排序
	Updater      int64     // 更新者ID
	UpdateDate   time.Time // 更新时间
	Creator      int64     // 创建者ID
	CreateDate   time.Time // 创建时间
}

// OtaResponseDTO OTA固件响应DTO（用于向前端展示，所有int64字段格式化为字符串）
type OtaResponseDTO struct {
	ID           string // ID
	FirmwareName string // 固件名称
	Type         string // 固件类型
	Version      string // 版本号
	Size         int64  // 文件大小(字节)
	Remark       string // 备注/说明
	FirmwarePath string // 固件路径
	Sort         int32  // 排序
	Updater      string // 更新者ID（格式化为字符串）
	UpdateDate   string // 更新时间（格式化）
	Creator      string // 创建者ID（格式化为字符串）
	CreateDate   string // 创建时间（格式化）
}

// ListOtaParams 分页查询参数
type ListOtaParams struct {
	FirmwareName *string // 可选，固件名称（模糊查询）
}

// OtaRepo OTA固件数据访问接口
type OtaRepo interface {
	// PageOta 分页查询OTA固件列表
	PageOta(ctx context.Context, params *ListOtaParams, page *kit.PageRequest) ([]*Ota, int, error)

	// GetByID 根据ID查询OTA固件记录
	GetByID(ctx context.Context, id string) (*Ota, error)

	// Save 保存OTA固件（同类固件只保留最新一条）
	Save(ctx context.Context, entity *Ota) error

	// Update 更新OTA固件（检查类型和版本唯一性）
	Update(ctx context.Context, entity *Ota) error

	// Delete 批量删除OTA固件
	Delete(ctx context.Context, ids []string) error
}

// OtaUsecase OTA固件业务逻辑
type OtaUsecase struct {
	repo        OtaRepo
	redisClient *kit.RedisClient
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewOtaUsecase 创建OTA固件用例
func NewOtaUsecase(
	repo OtaRepo,
	redisClient *kit.RedisClient,
	logger log.Logger,
) *OtaUsecase {
	return &OtaUsecase{
		repo:        repo,
		redisClient: redisClient,
		handleError: cerrors.NewHandleError(logger),
		log:         kit.LogHelper(logger),
	}
}

// PageOta 分页查询OTA固件列表
func (uc *OtaUsecase) PageOta(ctx context.Context, params *ListOtaParams, page *kit.PageRequest) ([]*OtaResponseDTO, int, error) {
	// 查询列表
	list, total, err := uc.repo.PageOta(ctx, params, page)
	if err != nil {
		return nil, 0, err
	}

	// 转换为DTO列表
	dtoList := make([]*OtaResponseDTO, 0, len(list))
	for _, entity := range list {
		dto := uc.convertToResponseDTO(entity)
		dtoList = append(dtoList, dto)
	}

	return dtoList, total, nil
}

// GetOtaByID 根据ID查询OTA固件信息
func (uc *OtaUsecase) GetOtaByID(ctx context.Context, id string) (*OtaResponseDTO, error) {
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("OTA固件记录不存在"))
	}

	return uc.convertToResponseDTO(entity), nil
}

// SaveOta 保存OTA固件
func (uc *OtaUsecase) SaveOta(ctx context.Context, entity *Ota) error {
	if entity == nil {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件信息不能为空"))
	}
	if entity.FirmwareName == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件名称不能为空"))
	}
	if entity.Type == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件类型不能为空"))
	}
	if entity.Version == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("版本号不能为空"))
	}

	if err := uc.repo.Save(ctx, entity); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// UpdateOta 更新OTA固件
func (uc *OtaUsecase) UpdateOta(ctx context.Context, id string, entity *Ota) error {
	if id == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}
	if entity == nil {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件信息不能为空"))
	}
	if entity.FirmwareName == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件名称不能为空"))
	}
	if entity.Type == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件类型不能为空"))
	}
	if entity.Version == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("版本号不能为空"))
	}

	entity.ID = id
	if err := uc.repo.Update(ctx, entity); err != nil {
		if err == ErrDuplicateOtaTypeVersion {
			return err
		}
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// DeleteOta 批量删除OTA固件
func (uc *OtaUsecase) DeleteOta(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("删除的固件ID不能为空"))
	}

	if err := uc.repo.Delete(ctx, ids); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// GetDownloadUrl 生成下载链接UUID并存储到Redis
func (uc *OtaUsecase) GetDownloadUrl(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	// 验证记录存在
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return "", uc.handleError.ErrInternal(ctx, err)
	}

	// 生成UUID
	uuidStr := uuid.New().String()

	// 存储到Redis（24小时过期）
	redisKey := kit.GetOtaIdKey(uuidStr)
	if err := uc.redisClient.Set(ctx, redisKey, id, 24*time.Hour); err != nil {
		return "", uc.handleError.ErrInternal(ctx, err)
	}

	return uuidStr, nil
}

// DownloadOta 下载OTA固件（从Redis获取ID，检查下载次数，返回文件数据）
func (uc *OtaUsecase) DownloadOta(ctx context.Context, uuidStr string) ([]byte, string, error) {
	if uuidStr == "" {
		return nil, "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("uuid不能为空"))
	}

	// 从Redis获取ID
	redisKey := kit.GetOtaIdKey(uuidStr)
	id, err := uc.redisClient.Get(ctx, redisKey)
	if err != nil || id == "" {
		return nil, "", uc.handleError.ErrNotFound(ctx, fmt.Errorf("下载链接不存在或已过期"))
	}

	// 检查下载次数
	downloadCountKey := kit.GetOtaDownloadCountKey(uuidStr)
	downloadCountStr, _ := uc.redisClient.Get(ctx, downloadCountKey)
	downloadCount := 0
	if downloadCountStr != "" {
		fmt.Sscanf(downloadCountStr, "%d", &downloadCount)
	}

	// 如果下载次数超过3次，删除Redis key并返回错误
	if downloadCount >= 3 {
		_ = uc.redisClient.Delete(ctx, downloadCountKey)
		_ = uc.redisClient.Delete(ctx, redisKey)
		return nil, "", uc.handleError.ErrNotFound(ctx, fmt.Errorf("下载次数已达上限"))
	}

	// 增加下载次数
	downloadCount++
	if err := uc.redisClient.Set(ctx, downloadCountKey, fmt.Sprintf("%d", downloadCount), 24*time.Hour); err != nil {
		return nil, "", uc.handleError.ErrInternal(ctx, err)
	}

	// 获取固件信息
	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil || entity.FirmwarePath == "" {
		return nil, "", uc.handleError.ErrNotFound(ctx, fmt.Errorf("固件文件不存在"))
	}

	// 读取文件
	fileData, filename, err := uc.readFirmwareFile(entity.FirmwarePath, entity.Type, entity.Version)
	if err != nil {
		return nil, "", uc.handleError.ErrInternal(ctx, err)
	}

	return fileData, filename, nil
}

// UploadFirmware 上传固件文件（计算MD5，保存文件，返回路径）
func (uc *OtaUsecase) UploadFirmware(ctx context.Context, fileData []byte, filename string) (string, error) {
	if len(fileData) == 0 {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("上传文件不能为空"))
	}
	if filename == "" {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("文件名不能为空"))
	}

	// 检查文件扩展名
	ext := filepath.Ext(filename)
	if ext != ".bin" && ext != ".apk" {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("只允许上传.bin和.apk格式的文件"))
	}

	// 计算MD5值
	md5Hash := md5.Sum(fileData)
	md5Str := fmt.Sprintf("%x", md5Hash)

	// 设置存储路径
	uploadDir := "uploadfile"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", uc.handleError.ErrInternal(ctx, fmt.Errorf("创建上传目录失败: %w", err))
	}

	// 使用MD5作为文件名，保留原始扩展名
	uniqueFileName := md5Str + ext
	filePath := filepath.Join(uploadDir, uniqueFileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件已存在，直接返回路径
		return filePath, nil
	}

	// 保存文件
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", uc.handleError.ErrInternal(ctx, fmt.Errorf("保存文件失败: %w", err))
	}

	return filePath, nil
}

// convertToResponseDTO 转换为响应DTO
func (uc *OtaUsecase) convertToResponseDTO(entity *Ota) *OtaResponseDTO {
	dto := &OtaResponseDTO{
		ID:           entity.ID,
		FirmwareName: entity.FirmwareName,
		Type:         entity.Type,
		Version:      entity.Version,
		Size:         entity.Size,
		Remark:       entity.Remark,
		FirmwarePath: entity.FirmwarePath,
		Sort:         entity.Sort,
	}

	// 格式化int64字段为字符串
	if entity.Creator > 0 {
		dto.Creator = fmt.Sprintf("%d", entity.Creator)
	}
	if entity.Updater > 0 {
		dto.Updater = fmt.Sprintf("%d", entity.Updater)
	}

	// 格式化时间字段
	if !entity.CreateDate.IsZero() {
		dto.CreateDate = entity.CreateDate.Format("2006-01-02 15:04:05")
	}
	if !entity.UpdateDate.IsZero() {
		dto.UpdateDate = entity.UpdateDate.Format("2006-01-02 15:04:05")
	}

	return dto
}

// readFirmwareFile 读取固件文件（支持绝对路径和相对路径查找）
func (uc *OtaUsecase) readFirmwareFile(firmwarePath, otaType, otaVersion string) ([]byte, string, error) {
	var filePath string
	var err error

	// 检查是否是绝对路径
	if filepath.IsAbs(firmwarePath) {
		filePath = firmwarePath
	} else {
		// 如果是相对路径，从当前工作目录解析
		wd, _ := os.Getwd()
		filePath = filepath.Join(wd, firmwarePath)
	}

	// 尝试读取文件
	fileData, err := os.ReadFile(filePath)
	if err == nil {
		// 文件读取成功，生成文件名
		filename := otaType + "_" + otaVersion
		if ext := filepath.Ext(firmwarePath); ext != "" {
			filename += ext
		}
		// 清理文件名，移除不安全字符
		filename = sanitizeFilename(filename)
		return fileData, filename, nil
	}

	// 如果文件不存在，尝试从firmware目录下查找文件名
	fileName := filepath.Base(firmwarePath)
	wd, _ := os.Getwd()
	altPath := filepath.Join(wd, "firmware", fileName)

	fileData, err = os.ReadFile(altPath)
	if err != nil {
		return nil, "", fmt.Errorf("固件文件不存在: %v", err)
	}

	// 文件读取成功，生成文件名
	filename := otaType + "_" + otaVersion
	if ext := filepath.Ext(fileName); ext != "" {
		filename += ext
	}
	// 清理文件名，移除不安全字符
	filename = sanitizeFilename(filename)
	return fileData, filename, nil
}

// sanitizeFilename 清理文件名，移除不安全字符
func sanitizeFilename(filename string) string {
	result := ""
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			result += string(r)
		} else {
			result += "_"
		}
	}
	return result
}

// calculateMD5 计算文件的MD5值（用于兼容旧代码，但当前使用内存中的文件数据）
func calculateMD5(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
