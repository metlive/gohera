package redis

import "errors"

// 有序集合命令

// ZAdd 向有序集合添加一个或多个成员，或更新已存在成员的分数
func (r *Client) ZAdd(key string, scores []int, members []string) (int, error) {
	if len(scores) != len(members) {
		return 0, errors.New("param error")
	}

	args := make([]any, 0, len(scores)*2+1)
	args = append(args, key)
	for i, score := range scores {
		args = append(args, score)
		args = append(args, members[i])
	}

	return r.int("ZADD", args...)
}

// ZCard 获取有序集合的成员数
func (r *Client) ZCard(key string) (int, error) {
	return r.int("ZCARD", key)
}

// ZCount 计算在有序集合中指定区间分数的成员数
func (r *Client) ZCount(key string, min, max int) (int, error) {
	return r.int("ZCOUNT", key, min, max)
}

// ZIncrBy 有序集合中对指定成员的分数加上增量
func (r *Client) ZIncrBy(key, member string, increment int) (int64, error) {
	return r.int64("ZINCRBY", key, increment, member)
}

// ZRange 通过索引区间返回有序集合指定区间内的成员
func (r *Client) ZRange(key string, start, stop int, withScores bool) ([]string, error) {
	args := make([]any, 0, 4)
	args = append(args, key)
	args = append(args, start)
	args = append(args, stop)
	if withScores {
		args = append(args, "WITHSCORES")
	}

	return r.strings("ZRANGE", args...)
}

// ZRangeByScore 通过分数返回有序集合指定区间内的成员
func (r *Client) ZRangeByScore(key string, min, max int, withScores bool) ([]string, error) {
	args := make([]any, 0, 4)
	args = append(args, key)
	args = append(args, min)
	args = append(args, max)
	if withScores {
		args = append(args, "WITHSCORES")
	}

	return r.strings("ZRANGEBYSCORE", args...)
}

// ZRank 返回有序集合中指定成员的索引
func (r *Client) ZRank(key, member string) (int, error) {
	return r.int("ZRANK", key, member)
}

// ZRem 移除有序集合中的一个或多个成员
func (r *Client) ZRem(key string, members ...any) (int, error) {
	args := make([]any, 0, len(members)+1)
	args = append(args, key)
	args = append(args, members...)

	return r.int("ZREM", args...)
}

// ZRemRangeByRank 移除有序集合中给定的排名区间的所有成员
func (r *Client) ZRemRangeByRank(key string, start, stop int) (int, error) {
	return r.int("ZREMRANGEBYRANK", key, start, stop)
}

// ZRemRangeByScore 移除有序集合中给定的分数区间的所有成员
func (r *Client) ZRemRangeByScore(key string, min, max int) (int, error) {
	return r.int("ZREMRANGEBYSCORE", key, min, max)
}

// ZRevRange 返回有序集合中指定区间内的成员，通过索引，分数从高到低
func (r *Client) ZRevRange(key string, start, stop int, withScores bool) ([]string, error) {
	args := make([]any, 0, 4)
	args = append(args, key)
	args = append(args, start)
	args = append(args, stop)
	if withScores {
		args = append(args, "WITHSCORES")
	}

	return r.strings("ZREVRANGE", args...)
}

// ZRevRangeByScore 返回有序集合中指定分数区间内的成员，分数从高到低
func (r *Client) ZRevRangeByScore(key string, max, min int, withScores bool) ([]string, error) {
	args := make([]any, 0, 4)
	args = append(args, key)
	args = append(args, max)
	args = append(args, min)
	if withScores {
		args = append(args, "WITHSCORES")
	}

	return r.strings("ZREVRANGEBYSCORE", args...)
}

// ZRevRank 返回有序集合中指定成员的排名，有序集合成员按分数值递减(从大到小)排序
func (r *Client) ZRevRank(key, member string) (int, error) {
	return r.int("ZREVRANK", key, member)
}

// ZScore 返回有序集合中，成员的分数值
func (r *Client) ZScore(key, member string) (int, error) {
	return r.int("ZSCORE", key, member)
}
