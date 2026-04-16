-- 用户资料扩展：昵称和加密手机号
ALTER TABLE users ADD COLUMN nickname VARCHAR(100) DEFAULT '';
ALTER TABLE users ADD COLUMN phone_encrypted TEXT DEFAULT '';
