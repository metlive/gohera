package redis

// 键操作命令

func (r *Client) Del(key string) (int, error) {
	return r.int("DEL", key)
}

func (r *Client) Exists(key string) (int, error) {
	return r.int("EXISTS", key)
}

func (r *Client) Expire(key string, seconds int) (int, error) {
	return r.int("EXPIRE", key, seconds)
}

func (r *Client) ExpireAt(key string, timestamp int) (int, error) {
	return r.int("EXPIREAT", key, timestamp)
}

func (r *Client) Ttl(key string) (int, error) {
	return r.int("TTL", key)
}

func (r *Client) Persist(key string) (int, error) {
	return r.int("PERSIST", key)
}

func (r *Client) PExpire(key string, milliseconds int64) (int, error) {
	return r.int("PEXPIRE", key, milliseconds)
}

func (r *Client) PExpireAt(key string, timestamp int64) (int, error) {
	return r.int("PEXPIREAT", key, timestamp)
}

func (r *Client) PTtl(key string) (int64, error) {
	return r.int64("PTTL", key)
}
