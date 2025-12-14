package redis

// 集合命令

// SAdd 向集合添加一个或多个成员
func (r *Client) SAdd(key string, members ...any) (int, error) {
	args := make([]any, 0, len(members)+1)
	args = append(args, key)
	args = append(args, members...)

	return r.int("SADD", args...)
}

// SCard 获取集合成员数
func (r *Client) SCard(key string) (int, error) {
	return r.int("SCARD", key)
}

// SMembers 返回集合中的所有成员
func (r *Client) SMembers(key string) ([]string, error) {
	return r.strings("SMEMBERS", key)
}

// SIsMember 判断成员元素是否是集合的成员
func (r *Client) SIsMember(key, member string) (int, error) {
	return r.int("SISMEMBER", key, member)
}

// SRem 移除集合中一个或多个成员
func (r *Client) SRem(key string, members ...any) (int, error) {
	args := make([]any, 0, len(members)+1)
	args = append(args, key)
	args = append(args, members...)

	return r.int("SREM", args...)
}

// SPop 移除并返回集合中的一个随机元素
func (r *Client) SPop(key string) (string, error) {
	return r.string("SPOP", key)
}

// SRandMember 返回集合中一个或多个随机数
func (r *Client) SRandMember(key string, count int) ([]string, error) {
	return r.strings("SRANDMEMBER", key, count)
}

// SMove 将 member 元素从 source 集合移动到 destination 集合
func (r *Client) SMove(source, destination, member string) (int, error) {
	return r.int("SMOVE", source, destination, member)
}

// SInter 返回给定所有集合的交集
func (r *Client) SInter(keys ...any) ([]string, error) {
	return r.strings("SINTER", keys...)
}

// SInterStore 返回给定所有集合的交集并存储在 destination 中
func (r *Client) SInterStore(destination string, keys ...any) (int, error) {
	args := make([]any, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)

	return r.int("SINTERSTORE", args...)
}

// SDiff 返回给定所有集合的差集
func (r *Client) SDiff(keys ...any) ([]string, error) {
	return r.strings("SDIFF", keys...)
}

// SDiffStore 返回给定所有集合的差集并存储在 destination 中
func (r *Client) SDiffStore(destination string, keys ...any) (int, error) {
	args := make([]any, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)

	return r.int("SDIFFSTORE", args...)
}

// SUnion 返回给定所有集合的并集
func (r *Client) SUnion(keys ...any) ([]string, error) {
	return r.strings("SUNION", keys...)
}

// SUnionStore 返回给定所有集合的并集并存储在 destination 中
func (r *Client) SUnionStore(destination string, keys ...any) (int, error) {
	args := make([]any, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)

	return r.int("SUNIONSTORE", args...)
}
