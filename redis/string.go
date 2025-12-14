package redis

// 字符串命令
// Set 设置字符串值
func (r *Client) Set(key, value string) (string, error) {
	return r.string("SET", key, value)
}

// Get 获取字符串值
func (r *Client) Get(key string) (string, error) {
	return r.string("GET", key)
}

// SetEx 设置字符串值并指定过期时间（秒）
func (r *Client) SetEx(key, value string, seconds int) (string, error) {
	return r.string("SETEX", key, seconds, value)
}

// SetNx 当键不存在时设置字符串值
func (r *Client) SetNx(key, value string) (int, error) {
	return r.int("SETNX", key, value)
}

// PSetEx 设置字符串值并指定过期时间（毫秒）
func (r *Client) PSetEx(key, value string, milliseconds int) (string, error) {
	return r.string("PSETEX", key, milliseconds, value)
}

// GetSet 设置新值并返回旧值
func (r *Client) GetSet(key, value string) (string, error) {
	return r.string("GETSET", key, value)
}

// GetRange 获取字符串的子串
func (r *Client) GetRange(key string, start, end int) (string, error) {
	return r.string("GETRANGE", key, start, end)
}

// Incr 将键的整数值加 1
func (r *Client) Incr(key string) (int64, error) {
	return r.int64("INCR", key)
}

// IncrBy 将键的整数值加上增量
func (r *Client) IncrBy(key string, increment int) (int64, error) {
	return r.int64("INCRBY", key, increment)
}

// Decr 将键的整数值减 1
func (r *Client) Decr(key string) (int64, error) {
	return r.int64("DECR", key)
}

// DecrBy 将键的整数值减去减量
func (r *Client) DecrBy(key string, decrement int) (int64, error) {
	return r.int64("DECRBY", key, decrement)
}

// Append 将值追加到键的末尾
func (r *Client) Append(key, value string) (int, error) {
	return r.int("APPEND", key, value)
}

// Strlen 获取字符串值的长度
func (r *Client) Strlen(key string) (int, error) {
	return r.int("STRLEN", key)
}
