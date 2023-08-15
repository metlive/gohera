package redis

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

// 分布式锁
// TODO 先简单实现，后续再优化
// 1. 延长锁

type DistributeLock struct {
	c            *Client
	expireSecond int
}

type DistributeLockOpt func(*DistributeLock)

// 初始化
func NewDistributeLock(driver *Client, opts ...DistributeLockOpt) *DistributeLock {
	d := &DistributeLock{
		c: driver,
	}

	for _, o := range opts {
		o(d)
	}

	return d
}

// 设置过期时间秒数
func SetExpireSecond(expire int) DistributeLockOpt {
	return func(d *DistributeLock) {
		d.expireSecond = expire
	}
}

// 加锁
func (d *DistributeLock) Lock(key, value string) (bool, error) {
	if d.expireSecond > 0 {
		reply, err := d.c.string("SET", key, value, "EX", d.expireSecond, "NX")
		if err != nil {
			return false, err
		}

		return reply == "OK", nil
	}

	return false, errors.New("expire second not set")
}

// lua脚本，用来释放分布式锁
var luaUnLock = "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end"

// 解锁
func (d *DistributeLock) UnLock(key, value string) (bool, error) {
	conn := d.c.pool.Get()
	defer conn.Close()

	lua := redis.NewScript(1, luaUnLock)
	r, err := redis.Int(lua.Do(conn, key, value))
	if err != nil {
		return false, err
	}

	if r > 0 {
		return true, nil
	} else {
		return false, nil
	}
}
