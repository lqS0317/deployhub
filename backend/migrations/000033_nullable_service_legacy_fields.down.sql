-- 恢复 NOT NULL（需要先填充默认值）
UPDATE services SET port = 8080 WHERE port = 0 OR port IS NULL;
UPDATE services SET image_repo = 'placeholder' WHERE image_repo = '' OR image_repo IS NULL;
ALTER TABLE services ALTER COLUMN port SET NOT NULL;
ALTER TABLE services ALTER COLUMN image_repo SET NOT NULL;
