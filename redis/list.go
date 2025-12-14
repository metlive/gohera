package redis

// 列表命令

// LPush 将一个或多个值插入到列表头部
func (r *Client) LPush(key string, values ...any) (int, error) {
	args := make([]any, 0, len(values)+1)
	args = append(args, key)
	args = append(args, values...)

	return r.int("LPUSH", args...)
}

// RPush 将一个或多个值插入到列表尾部
func (r *Client) RPush(key string, values ...any) (int, error) {
	args := make([]any, 0, len(values)+1)
	args = append(args, key)
	args = append(args, values...)

	return r.int("RPUSH", args...)
}

// LPushX 当列表存在时，将值插入到列表头部
func (r *Client) LPushX(key string, value string) (int, error) {
	return r.int("LPUSHX", key, value)
}

// RPushX 当列表存在时，将值插入到列表尾部
func (r *Client) RPushX(key string, value string) (int, error) {
	return r.int("RPUSHX", key, value)
}

// LPop 移除并返回列表的第一个元素
func (r *Client) LPop(key string) (string, error) {
	return r.string("LPOP", key)
}

// RPop 移除并返回列表的最后一个元素
func (r *Client) RPop(key string) (string, error) {
	return r.string("RPOP", key)
}

// RPopLPush 移除列表最后一个元素，并将该元素添加到另一个列表头部
func (r *Client) RPopLPush(source, destination string) (string, error) {
	return r.string("RPOPLPUSH", source, destination)
}

// LLen 获取列表长度
func (r *Client) LLen(key string) (int, error) {
	return r.int("LLEN", key)
}

// LIndex 通过索引获取列表中的元素
func (r *Client) LIndex(key string, index int) (string, error) {
	return r.string("LINDEX", key, index)
}

// LRange 获取列表指定范围内的元素
func (r *Client) LRange(key string, start, stop int) ([]string, error) {
	return r.strings("LRANGE", key, start, stop)
}

// LRem 移除列表元素
func (r *Client) LRem(key string, count int, value string) (int, error) {
	return r.int("LREM", key, count, value)
}

// LSet 通过索引设置列表元素的值
func (r *Client) LSet(key string, index int, value string) (bool, error) {
	reply, err := r.string("LSET", key, index, value)
	if err != nil {
		return false, err
	}

	return reply == "OK", nil
}
