package redis

// 哈希表命令

// HSet 设置哈希表中字段的值
func (r *Client) HSet(key, field, value string) (int, error) {
	return r.int("HSET", key, field, value)
}

// HSetNx 当字段不存在时设置哈希表字段的值
func (r *Client) HSetNx(key, field, value string) (int, error) {
	return r.int("HSETNX", key, field, value)
}

// HGet 获取哈希表中字段的值
func (r *Client) HGet(key, field string) (string, error) {
	return r.string("HGET", key, field)
}

// HGetAll 获取哈希表中的所有字段和值
func (r *Client) HGetAll(key string) (map[string]string, error) {
	return r.stringMap("HGETALL", key)
}

// HExists 检查哈希表中字段是否存在
func (r *Client) HExists(key string) (map[string]string, error) {
	return r.stringMap("HEXISTS", key)
}

// HDel 删除哈希表中的一个或多个字段
func (r *Client) HDel(key string, fields ...any) (int, error) {
	return r.int("HDEL", key, fields)
}

// HKeys 获取哈希表中的所有字段名
func (r *Client) HKeys(key string) ([]string, error) {
	return r.strings("HKEYS", key)
}

// HVals 获取哈希表中的所有值
func (r *Client) HVals(key string) ([]string, error) {
	return r.strings("HVALS", key)
}

// HLen 获取哈希表中字段的数量
func (r *Client) HLen(key string) (int, error) {
	return r.int("HLEN", key)
}

// HIncrBy 为哈希表字段的整数值加上增量
func (r *Client) HIncrBy(key, field string, increment int) (int64, error) {
	return r.int64("HINCRBY", key, field, increment)
}
