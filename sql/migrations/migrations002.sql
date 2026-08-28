-- 增量迁移：movies 补齐 genre / release_date / rating（与 domain.Movie、MovieRepo 对齐）

ALTER TABLE movies ADD COLUMN IF NOT EXISTS genre        VARCHAR(64);
ALTER TABLE movies ADD COLUMN IF NOT EXISTS release_date DATE;
ALTER TABLE movies ADD COLUMN IF NOT EXISTS rating       NUMERIC(3,1);
