-- 添加逻辑操作符字段到AI策略表
-- 支持 AND（同时满足）、OR（满足其中一个）、NOT（都不满足）

ALTER TABLE ai_control_strategies
ADD COLUMN logic_operator VARCHAR(10) DEFAULT 'AND';

-- 添加列注释
COMMENT ON COLUMN ai_control_strategies.logic_operator IS '条件逻辑操作符: AND, OR, NOT';

-- 更新现有记录的默认值
UPDATE ai_control_strategies
SET logic_operator = 'AND'
WHERE logic_operator IS NULL OR logic_operator = '';

-- 添加检查约束确保只能使用有效的操作符
ALTER TABLE ai_control_strategies
ADD CONSTRAINT chk_logic_operator
CHECK (logic_operator IN ('AND', 'OR', 'NOT'));
