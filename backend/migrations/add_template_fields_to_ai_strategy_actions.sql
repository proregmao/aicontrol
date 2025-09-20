-- 为ai_strategy_actions表添加动作模板支持字段
-- 执行时间: 2025-09-20

-- 添加模板相关字段
ALTER TABLE ai_strategy_actions 
ADD COLUMN template_id INTEGER,
ADD COLUMN template_name VARCHAR(100),
ADD COLUMN use_template BOOLEAN DEFAULT FALSE;

-- 添加外键约束（如果需要的话）
-- ALTER TABLE ai_strategy_actions 
-- ADD CONSTRAINT fk_ai_strategy_actions_template_id 
-- FOREIGN KEY (template_id) REFERENCES action_templates(id) ON DELETE SET NULL;

-- 添加索引以提高查询性能
CREATE INDEX idx_ai_strategy_actions_template_id ON ai_strategy_actions(template_id);
CREATE INDEX idx_ai_strategy_actions_use_template ON ai_strategy_actions(use_template);

-- 添加注释
COMMENT ON COLUMN ai_strategy_actions.template_id IS '动作模板ID，关联action_templates表';
COMMENT ON COLUMN ai_strategy_actions.template_name IS '动作模板名称，用于显示';
COMMENT ON COLUMN ai_strategy_actions.use_template IS '是否使用动作模板';
