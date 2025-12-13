package redis

// 哈希表命令

func (r *Client) HSet(key, field, value string) (int, error) {
	return r.int("HSET", key, field, value)
}

func (r *Client) HSetNx(key, field, value string) (int, error) {
	return r.int("HSETNX", key, field, value)
}

func (r *Client) HGet(key, field string) (string, error) {
	return r.string("HGET", key, field)
}

func (r *Client) HGetAll(key string) (map[string]string, error) {
	return r.stringMap("HGETALL", key)
}

func (r *Client) HExists(key string) (map[string]string, error) {
	return r.stringMap("HEXISTS", key)
}

func (r *Client) HDel(key string, fields ...any) (int, error) {
	return r.int("HDEL", key, fields)
}

func (r *Client) HKeys(key string) ([]string, error) {
	return r.strings("HKEYS", key)
}

func (r *Client) HVals(key string) ([]string, error) {
	return r.strings("HVALS", key)
}

func (r *Client) HLen(key string) (int, error) {
	return r.int("HLEN", key)
}

func (r *Client) HIncrBy(key, field string, increment int) (int64, error) {
	return r.int64("HINCRBY", key, field, increment)
}
