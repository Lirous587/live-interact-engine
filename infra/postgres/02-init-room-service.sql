-- ==================== Room Service 数据库初始化 ====================

-- 创建 rooms 表
CREATE TABLE IF NOT EXISTS rooms (
    room_id VARCHAR(100) PRIMARY KEY,
    owner_id VARCHAR(100) NOT NULL,
    title VARCHAR(30) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_rooms_is_active ON rooms(is_active);

-- 创建 user_room_roles 表
CREATE TABLE IF NOT EXISTS user_room_roles (
    user_id VARCHAR(100) NOT NULL,
    room_id VARCHAR(100) NOT NULL,
    role_name VARCHAR(50) NOT NULL,
    permissions JSONB DEFAULT '[]',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    PRIMARY KEY (user_id, room_id),
    FOREIGN KEY (room_id) REFERENCES rooms(room_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_room_roles_room_id ON user_room_roles(room_id);
CREATE INDEX IF NOT EXISTS idx_user_room_roles_user_id ON user_room_roles(user_id);
