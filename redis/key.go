package redis

// 键操作命令

// Del 删除指定的键
func (r *Client) Del(key string) (int, error) {
	return r.int("DEL", key)
}

// Exists 检查键是否存在
func (r *Client) Exists(key string) (int, error) {
	return r.int("EXISTS", key)
}

// Expire 设置键的过期时间（秒）
func (r *Client) Expire(key string, seconds int) (int, error) {
	return r.int("EXPIRE", key, seconds)
}

// ExpireAt 设置键在指定时间戳过期（秒）
func (r *Client) ExpireAt(key string, timestamp int) (int, error) {
	return r.int("EXPIREAT", key, timestamp)
}

// Ttl 获取键的剩余生存时间（秒）
func (r *Client) Ttl(key string) (int, error) {
	return r.int("TTL", key)
}

// Persist 移除键的过期时间
func (r *Client) Persist(key string) (int, error) {
	return r.int("PERSIST", key)
}

// PExpire 设置键的过期时间（毫秒）
func (r *Client) PExpire(key string, milliseconds int64) (int, error) {
	return r.int("PEXPIRE", key, milliseconds)
}

// PExpireAt 设置键在指定时间戳过期（毫秒）
func (r *Client) PExpireAt(key string, timestamp int64) (int, error) {
	return r.int("PEXPIREAT", key, timestamp)
}

// PTtl 获取键的剩余生存时间（毫秒）
func (r *Client) PTtl(key string) (int64, error) {
	return r.int64("PTTL", key)
}
