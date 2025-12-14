package redis

import (
	"github.com/gomodule/redigo/redis"
)

// Lock 获取分布式锁
// key: 锁的键名
// requestId: 请求标识（建议使用 UUID），用于解锁时校验锁的持有者，防止误删
// ttl: 锁的过期时间（秒），防止死锁
// 返回值: true 表示加锁成功，false 表示锁已被占用
func (r *Client) Lock(key, requestId string, ttl int) (bool, error) {
	conn := r.pool.Get()
	defer conn.Close()

	// 使用 SET key value NX EX ttl 命令
	// NX: 仅在键不存在时设置
	// EX: 设置过期时间（秒）
	result, err := redis.String(conn.Do("SET", key, requestId, "NX", "EX", ttl))

	if err == redis.ErrNil {
		// 锁已被占用
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "OK", nil
}

// Unlock 释放分布式锁
// key: 锁的键名
// requestId: 加锁时使用的请求标识，必须匹配才能释放锁
// 返回值: true 表示释放成功，false 表示锁不存在或不属于该 requestId
func (r *Client) Unlock(key, requestId string) (bool, error) {
	conn := r.pool.Get()
	defer conn.Close()

	// 使用 Lua 脚本保证原子性：仅当 key 存在且 value 等于 requestId 时才删除
	script := redis.NewScript(1, `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)

	res, err := redis.Int(script.Do(conn, key, requestId))
	if err != nil {
		return false, err
	}

	// 返回 1 表示删除成功
	return res == 1, nil
}
