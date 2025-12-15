package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/biz/voice_print_client"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// 声纹识别阈值
const RECOGNITION_THRESHOLD = 0.5

// AgentVoicePrint 智能体声纹业务实体
type AgentVoicePrint struct {
	ID         string
	AgentID    string
	AudioID    string
	SourceName string
	Introduce  string
	CreateDate time.Time
	Creator    int64
	UpdateDate time.Time
	Updater    int64
}

// AgentVoicePrintVO 智能体声纹视图对象
type AgentVoicePrintVO struct {
	ID         string
	AudioID    string
	SourceName string
	Introduce  string
	CreateDate time.Time
}

// AgentVoicePrintRepo 智能体声纹数据访问接口
type AgentVoicePrintRepo interface {
	CreateAgentVoicePrint(ctx context.Context, voicePrint *AgentVoicePrint) (*AgentVoicePrint, error)
	UpdateAgentVoicePrint(ctx context.Context, voicePrint *AgentVoicePrint) error
	DeleteAgentVoicePrint(ctx context.Context, id string, userId int64) error
	ListAgentVoicePrints(ctx context.Context, agentId string, userId int64) ([]*AgentVoicePrint, error)
	GetAgentVoicePrintByID(ctx context.Context, id string) (*AgentVoicePrint, error)
	ListAgentVoicePrintIDsByAgentID(ctx context.Context, agentId string) ([]string, error)
	// ListAgentVoicePrintsByAgentID 根据智能体ID获取声纹列表（不需要用户权限验证）
	ListAgentVoicePrintsByAgentID(ctx context.Context, agentId string) ([]*AgentVoicePrint, error)
}

// AgentVoicePrintUsecase 智能体声纹业务逻辑
type AgentVoicePrintUsecase struct {
	repo             AgentVoicePrintRepo
	agentRepo        AgentRepo
	configUsecase    *ConfigUsecase
	voicePrintClient *voice_print_client.VoicePrintClient
	handleError      *cerrors.HandleError
	log              *log.Helper
}

// NewAgentVoicePrintUsecase 创建智能体声纹用例
func NewAgentVoicePrintUsecase(
	repo AgentVoicePrintRepo,
	agentRepo AgentRepo,
	configUsecase *ConfigUsecase,
	logger log.Logger,
) *AgentVoicePrintUsecase {
	return &AgentVoicePrintUsecase{
		repo:             repo,
		agentRepo:        agentRepo,
		configUsecase:    configUsecase,
		voicePrintClient: voice_print_client.NewVoicePrintClient(logger),
		handleError:      cerrors.NewHandleError(logger),
		log:              kit.LogHelper(logger),
	}
}

// getVoicePrintURL 获取声纹服务地址
func (uc *AgentVoicePrintUsecase) getVoicePrintURL(ctx context.Context) (string, error) {
	url, err := uc.configUsecase.GetValue(ctx, "server.voice_print", false)
	if err != nil {
		return "", fmt.Errorf("获取声纹接口地址失败: %w", err)
	}
	if url == "" || url == "null" {
		return "", fmt.Errorf("声纹接口未配置，请在参数配置中配置声纹接口地址(server.voice_print)")
	}
	return url, nil
}

// getVoicePrintAudioWAV 获取声纹音频数据
func (uc *AgentVoicePrintUsecase) getVoicePrintAudioWAV(ctx context.Context, agentId, audioId string) ([]byte, error) {
	// 判断这个音频是否属于当前智能体
	owned, err := uc.agentRepo.IsAudioOwnedByAgent(ctx, audioId, agentId)
	if err != nil {
		return nil, fmt.Errorf("检查音频归属失败: %w", err)
	}
	if !owned {
		return nil, fmt.Errorf("音频数据不属于该智能体")
	}

	// 获取音频数据
	audio, err := uc.agentRepo.GetAudioByID(ctx, audioId)
	if err != nil {
		return nil, fmt.Errorf("获取音频数据失败: %w", err)
	}

	// 如果音频数据为空直接报错
	if len(audio) == 0 {
		return nil, fmt.Errorf("音频数据为空")
	}

	return audio, nil
}

// CreateAgentVoicePrint 创建智能体声纹
func (uc *AgentVoicePrintUsecase) CreateAgentVoicePrint(ctx context.Context, agentId, audioId, sourceName string, introduce *string, userId int64) error {
	// 获取音频数据
	audioData, err := uc.getVoicePrintAudioWAV(ctx, agentId, audioId)
	if err != nil {
		return err
	}

	// 获取声纹服务地址
	voicePrintURL, err := uc.getVoicePrintURL(ctx)
	if err != nil {
		return err
	}

	// 识别一下此声音是否注册过
	existingIDs, err := uc.repo.ListAgentVoicePrintIDsByAgentID(ctx, agentId)
	if err != nil {
		return fmt.Errorf("查询现有声纹失败: %w", err)
	}

	if len(existingIDs) > 0 {
		response, err := uc.voicePrintClient.IdentifyVoicePrint(voicePrintURL, existingIDs, audioData)
		if err != nil {
			return fmt.Errorf("声纹识别失败: %w", err)
		}
		if response != nil && response.Score > RECOGNITION_THRESHOLD {
			// 根据识别出的声纹ID查询对应的用户信息
			existingVoicePrint, err := uc.repo.GetAgentVoicePrintByID(ctx, response.SpeakerID)
			existingUserName := "未知用户"
			if err == nil && existingVoicePrint != nil {
				existingUserName = existingVoicePrint.SourceName
			}
			return fmt.Errorf("此声音声纹已经注册，属于：%s", existingUserName)
		}
	}

	// 生成声纹ID（UUID去掉横线）
	voicePrintID := strings.ReplaceAll(uuid.New().String(), "-", "")

	// 创建声纹实体
	introduceStr := ""
	if introduce != nil {
		introduceStr = *introduce
	}
	voicePrint := &AgentVoicePrint{
		ID:         voicePrintID,
		AgentID:    agentId,
		AudioID:    audioId,
		SourceName: sourceName,
		Introduce:  introduceStr,
		CreateDate: time.Now(),
		Creator:    userId,
		UpdateDate: time.Now(),
		Updater:    userId,
	}

	// 开启事务：先保存数据库，再调用HTTP API
	// 注意：这里使用ent的事务，但为了简化，我们先保存数据库，如果HTTP失败则删除
	created, err := uc.repo.CreateAgentVoicePrint(ctx, voicePrint)
	if err != nil {
		return fmt.Errorf("保存声纹信息失败: %w", err)
	}

	// 发送注册声纹请求
	if err := uc.voicePrintClient.RegisterVoicePrint(voicePrintURL, created.ID, audioData); err != nil {
		// HTTP注册失败，删除已保存的数据库记录
		_ = uc.repo.DeleteAgentVoicePrint(ctx, created.ID, userId)
		return fmt.Errorf("注册声纹到外部服务失败: %w", err)
	}

	return nil
}

// UpdateAgentVoicePrint 更新智能体声纹
func (uc *AgentVoicePrintUsecase) UpdateAgentVoicePrint(ctx context.Context, id string, audioId, sourceName *string, introduce *string, userId int64) error {
	// 查询现有声纹
	existingVoicePrint, err := uc.repo.GetAgentVoicePrintByID(ctx, id)
	if err != nil {
		return fmt.Errorf("声纹不存在: %w", err)
	}

	// 检查权限
	if existingVoicePrint.Creator != userId {
		return fmt.Errorf("无权修改此声纹")
	}

	agentId := existingVoicePrint.AgentID
	var audioData []byte
	needReregister := false

	// 如果提供了audioId且与原来的不同，需要重新获取音频并注册
	if audioId != nil && *audioId != "" && *audioId != existingVoicePrint.AudioID {
		needReregister = true
		audioData, err = uc.getVoicePrintAudioWAV(ctx, agentId, *audioId)
		if err != nil {
			return err
		}

		// 获取声纹服务地址
		voicePrintURL, err := uc.getVoicePrintURL(ctx)
		if err != nil {
			return err
		}

		// 识别一下此声音是否注册过
		existingIDs, err := uc.repo.ListAgentVoicePrintIDsByAgentID(ctx, agentId)
		if err != nil {
			return fmt.Errorf("查询现有声纹失败: %w", err)
		}

		if len(existingIDs) > 0 {
			response, err := uc.voicePrintClient.IdentifyVoicePrint(voicePrintURL, existingIDs, audioData)
			if err != nil {
				return fmt.Errorf("声纹识别失败: %w", err)
			}
			// 返回分数高于阈值说明这个声纹已经有了
			if response != nil && response.Score > RECOGNITION_THRESHOLD {
				// 判断返回的id如果不是要修改的声纹id，说明这个声纹id，现在要注册的声音已经存在且不是原来的声纹，不允许修改
				if response.SpeakerID != id {
					// 根据识别出的声纹ID查询对应的用户信息
					existingVoicePrint2, err := uc.repo.GetAgentVoicePrintByID(ctx, response.SpeakerID)
					existingUserName := "未知用户"
					if err == nil && existingVoicePrint2 != nil {
						existingUserName = existingVoicePrint2.SourceName
					}
					return fmt.Errorf("声纹修改不允许，此声音已注册为声纹（%s）", existingUserName)
				}
			}
		}
	}

	// 更新声纹实体
	updateVoicePrint := &AgentVoicePrint{
		ID:         id,
		AgentID:    agentId,
		AudioID:    existingVoicePrint.AudioID,
		SourceName: existingVoicePrint.SourceName,
		Introduce:  existingVoicePrint.Introduce,
		UpdateDate: time.Now(),
		Updater:    userId,
	}

	if audioId != nil && *audioId != "" {
		updateVoicePrint.AudioID = *audioId
	}
	if sourceName != nil && *sourceName != "" {
		updateVoicePrint.SourceName = *sourceName
	}
	if introduce != nil {
		updateVoicePrint.Introduce = *introduce
	}

	// 更新数据库
	if err := uc.repo.UpdateAgentVoicePrint(ctx, updateVoicePrint); err != nil {
		return fmt.Errorf("更新声纹信息失败: %w", err)
	}

	// 如果需要重新注册
	if needReregister {
		voicePrintURL, err := uc.getVoicePrintURL(ctx)
		if err != nil {
			return err
		}

		// 先注销之前这个声纹id上的声纹向量
		if err := uc.voicePrintClient.CancelVoicePrint(voicePrintURL, id); err != nil {
			uc.log.Warnf("注销旧声纹失败: %v", err)
			// 继续执行，不中断流程
		}

		// 发送注册声纹请求
		if err := uc.voicePrintClient.RegisterVoicePrint(voicePrintURL, id, audioData); err != nil {
			return fmt.Errorf("重新注册声纹到外部服务失败: %w", err)
		}
	}

	return nil
}

// DeleteAgentVoicePrint 删除智能体声纹
func (uc *AgentVoicePrintUsecase) DeleteAgentVoicePrint(ctx context.Context, id string, userId int64) error {
	// 删除数据库记录
	if err := uc.repo.DeleteAgentVoicePrint(ctx, id, userId); err != nil {
		return fmt.Errorf("删除声纹失败: %w", err)
	}

	// 数据库删除成功后，异步调用注销API
	go func() {
		voicePrintURL, err := uc.getVoicePrintURL(context.Background())
		if err != nil {
			uc.log.Errorf("获取声纹接口地址失败: %v", err)
			return
		}

		if err := uc.voicePrintClient.CancelVoicePrint(voicePrintURL, id); err != nil {
			uc.log.Errorf("删除声纹存在运行时错误原因：%v，id：%s", err, id)
		}
	}()

	return nil
}

// ListAgentVoicePrints 获取智能体声纹列表
func (uc *AgentVoicePrintUsecase) ListAgentVoicePrints(ctx context.Context, agentId string, userId int64) ([]*AgentVoicePrintVO, error) {
	// 检查声纹接口是否配置
	_, err := uc.getVoicePrintURL(ctx)
	if err != nil {
		return nil, err
	}

	// 查询声纹列表
	voicePrints, err := uc.repo.ListAgentVoicePrints(ctx, agentId, userId)
	if err != nil {
		return nil, fmt.Errorf("查询声纹列表失败: %w", err)
	}

	// 转换为VO列表
	voList := make([]*AgentVoicePrintVO, 0, len(voicePrints))
	for _, vp := range voicePrints {
		voList = append(voList, &AgentVoicePrintVO{
			ID:         vp.ID,
			AudioID:    vp.AudioID,
			SourceName: vp.SourceName,
			Introduce:  vp.Introduce,
			CreateDate: vp.CreateDate,
		})
	}

	return voList, nil
}
