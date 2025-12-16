-- Daily Fetch PR 迁移：新增系统参数
-- 执行时间：2025-12-16

-- 1. 删除并重新添加 server.auth.enabled 参数
DELETE FROM `sys_params` WHERE param_code = 'server.auth.enabled';

INSERT INTO `sys_params` (id, param_code, param_value, value_type, param_type, remark) VALUES 
(122, 'server.auth.enabled', 'true', 'boolean', 1, 'server模块是否开启token认证');

-- 2. 删除并重新添加 system-web.menu 参数
DELETE FROM `sys_params` WHERE param_code = 'system-web.menu';

INSERT INTO `sys_params` (id, param_code, param_value, value_type, param_type, remark) VALUES 
(600, 'system-web.menu', '{"features":{"voiceprintRecognition":{"name":"feature.voiceprintRecognition.name","enabled":false,"description":"feature.voiceprintRecognition.description"},"voiceClone":{"name":"feature.voiceClone.name","enabled":false,"description":"feature.voiceClone.description"},"knowledgeBase":{"name":"feature.knowledgeBase.name","enabled":false,"description":"feature.knowledgeBase.description"},"mcpAccessPoint":{"name":"feature.mcpAccessPoint.name","enabled":false,"description":"feature.mcpAccessPoint.description"},"vad":{"name":"feature.vad.name","enabled":true,"description":"feature.vad.description"},"asr":{"name":"feature.asr.name","enabled":true,"description":"feature.asr.description"}},"groups":{"featureManagement":["voiceprintRecognition","voiceClone","knowledgeBase","mcpAccessPoint"],"voiceManagement":["vad","asr"]}}', 'json', 1, '系统功能菜单配置');

-- 3. 创建 ai_agent_context_provider 表
CREATE TABLE IF NOT EXISTS `ai_agent_context_provider` (
    `id` VARCHAR(32) NOT NULL COMMENT '主键',
    `agent_id` VARCHAR(32) NOT NULL COMMENT '智能体ID',
    `context_providers` JSON COMMENT '上下文源配置',
    `creator` BIGINT COMMENT '创建者',
    `created_at` DATETIME COMMENT '创建时间',
    `updater` BIGINT COMMENT '更新者',
    `updated_at` DATETIME COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_agent_id` (`agent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='智能体上下文源配置表';
