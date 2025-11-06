-- 添加邮箱验证相关字段到用户表
ALTER TABLE users 
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN email_verified_at TIMESTAMP NULL;

-- 创建邮箱验证表
CREATE TABLE email_verifications (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NULL,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_token (token),
    INDEX idx_email (email),
    INDEX idx_expires_at (expires_at)
);

-- 创建索引以优化查询性能
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_users_email_verified_at ON users(email_verified_at);

-- 更新updated_at触发器
DELIMITER $$
CREATE TRIGGER update_email_verifications_updated_at 
    BEFORE UPDATE ON email_verifications
    FOR EACH ROW 
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$
DELIMITER ;
