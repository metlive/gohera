package redis

// 字符串命令
func (r *Client) Set(key, value string) (string, error) {
	return r.string("SET", key, value)
}

func (r *Client) Get(key string) (string, error) {
	return r.string("GET", key)
}

func (r *Client) SetEx(key, value string, seconds int) (string, error) {
	return r.string("SETEX", key, seconds, value)
}

func (r *Client) SetNx(key, value string) (int, error) {
	return r.int("SETNX", key, value)
}

func (r *Client) PSetEx(key, value string, milliseconds int) (string, error) {
	return r.string("PSETEX", key, milliseconds, value)
}

func (r *Client) GetSet(key, value string) (string, error) {
	return r.string("GETSET", key, value)
}

func (r *Client) GetRange(key string, start, end int) (string, error) {
	return r.string("GETRANGE", key, start, end)
}

func (r *Client) Incr(key string) (int64, error) {
	return r.int64("INCR", key)
}

func (r *Client) IncrBy(key string, increment int) (int64, error) {
	return r.int64("INCRBY", key, increment)
}

func (r *Client) Decr(key string) (int64, error) {
	return r.int64("DECR", key)
}

func (r *Client) DecrBy(key string, decrement int) (int64, error) {
	return r.int64("DECRBY", key, decrement)
}

func (r *Client) Append(key, value string) (int, error) {
	return r.int("APPEND", key, value)
}

func (r *Client) Strlen(key string) (int, error) {
	return r.int("STRLEN", key)
}
