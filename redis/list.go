package redis

// 列表命令

func (r *Client) LPush(key string, values ...any) (int, error) {
	args := make([]any, 0, len(values)+1)
	args = append(args, key)
	args = append(args, values...)

	return r.int("LPUSH", args...)
}

func (r *Client) RPush(key string, values ...any) (int, error) {
	args := make([]any, 0, len(values)+1)
	args = append(args, key)
	args = append(args, values...)

	return r.int("RPUSH", args...)
}

func (r *Client) LPushX(key string, value string) (int, error) {
	return r.int("LPUSHX", key, value)
}

func (r *Client) RPushX(key string, value string) (int, error) {
	return r.int("RPUSHX", key, value)
}

func (r *Client) LPop(key string) (string, error) {
	return r.string("LPOP", key)
}

func (r *Client) RPop(key string) (string, error) {
	return r.string("RPOP", key)
}

func (r *Client) RPopLPush(source, destination string) (string, error) {
	return r.string("RPOPLPUSH", source, destination)
}

func (r *Client) LLen(key string) (int, error) {
	return r.int("LLEN", key)
}

func (r *Client) LIndex(key string, index int) (string, error) {
	return r.string("LINDEX", key, index)
}

func (r *Client) LRange(key string, start, stop int) ([]string, error) {
	return r.strings("LRANGE", key, start, stop)
}

func (r *Client) LRem(key string, count int, value string) (int, error) {
	return r.int("LREM", key, count, value)
}

func (r *Client) LSet(key string, index int, value string) (bool, error) {
	reply, err := r.string("LSET", key, index, value)
	if err != nil {
		return false, err
	}

	return reply == "OK", nil
}
