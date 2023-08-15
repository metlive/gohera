package redis

// 集合命令

func (r *Client) SAdd(key string, members ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, key)
	args = append(args, members...)

	return r.int("SADD", args...)
}

func (r *Client) SCard(key string) (int, error) {
	return r.int("SCARD", key)
}

func (r *Client) SMembers(key string) ([]string, error) {
	return r.strings("SMEMBERS", key)
}

func (r *Client) SIsMember(key, member string) (int, error) {
	return r.int("SISMEMBER", key, member)
}

func (r *Client) SRem(key string, members ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, key)
	args = append(args, members...)

	return r.int("SREM", args...)
}

func (r *Client) SPop(key string) (string, error) {
	return r.string("SPOP", key)
}

func (r *Client) SRandMember(key string, count int) ([]string, error) {
	return r.strings("SRANDMEMBER", key, count)
}

func (r *Client) SMove(source, destination, member string) (int, error) {
	return r.int("SMOVE", source, destination, member)
}

func (r *Client) SInter(keys ...interface{}) ([]string, error) {
	return r.strings("SINTER", keys...)
}

func (r *Client) SInterStore(destination string, keys ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)

	return r.int("SINTERSTORE", args...)
}

func (r *Client) SDiff(keys ...interface{}) ([]string, error) {
	return r.strings("SDIFF", keys...)
}

func (r *Client) SDiffStore(destination string, keys ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)

	return r.int("SDIFFSTORE", args...)
}

func (r *Client) SUnion(keys ...interface{}) ([]string, error) {
	return r.strings("SUNION", keys...)
}

func (r *Client) SUnionStore(destination string, keys ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)

	return r.int("SUNIONSTORE", args...)
}
