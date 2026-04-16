-- env 列从 VARCHAR(10) 扩展到 VARCHAR(20)，以容纳 "development" 等值
ALTER TABLE clusters ALTER COLUMN env TYPE VARCHAR(20);
