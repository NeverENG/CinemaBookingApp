-- 增量迁移：users 补齐 total_reclaimed_points（与 postgres.UserRow 对齐）

ALTER TABLE users ADD COLUMN IF NOT EXISTS total_reclaimed_points INT NOT NULL DEFAULT 0;
