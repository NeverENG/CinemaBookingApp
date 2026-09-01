-- 场次票价规则（兼容已存在数据库）

ALTER TABLE show_sessions
    ADD COLUMN IF NOT EXISTS price_rules_json JSONB NOT NULL DEFAULT '{}';
