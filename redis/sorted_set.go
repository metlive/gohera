package redis

import "errors"

// 有序集合命令

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

func (r *Client) ZCard(key string) (int, error) {
	return r.int("ZCARD", key)
}

func (r *Client) ZCount(key string, min, max int) (int, error) {
	return r.int("ZCOUNT", key, min, max)
}

func (r *Client) ZIncrBy(key, member string, increment int) (int64, error) {
	return r.int64("ZINCRBY", key, increment, member)
}

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

func (r *Client) ZRank(key, member string) (int, error) {
	return r.int("ZRANK", key, member)
}

func (r *Client) ZRem(key string, members ...any) (int, error) {
	args := make([]any, 0, len(members)+1)
	args = append(args, key)
	args = append(args, members...)

	return r.int("ZREM", args...)
}

func (r *Client) ZRemRangeByRank(key string, start, stop int) (int, error) {
	return r.int("ZREMRANGEBYRANK", key, start, stop)
}

func (r *Client) ZRemRangeByScore(key string, min, max int) (int, error) {
	return r.int("ZREMRANGEBYSCORE", key, min, max)
}

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

func (r *Client) ZRevRank(key, member string) (int, error) {
	return r.int("ZREVRANK", key, member)
}

func (r *Client) ZScore(key, member string) (int, error) {
	return r.int("ZSCORE", key, member)
}
