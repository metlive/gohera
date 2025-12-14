package redis

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// RateLimit 令牌桶限流器
// key: 限流资源的键名
// rate: 令牌生成速率 (每秒生成的令牌数量)
// capacity: 桶的最大容量 (允许的最大突发请求数)
// required: 本次请求需要消耗的令牌数量 (通常为 1)
// 返回值: true 表示允许通过，false 表示被限流
func (r *Client) RateLimit(key string, rate int, capacity int, required int) (bool, error) {
	conn := r.pool.Get()
	defer conn.Close()

	// 获取当前时间（微秒），用于高精度计算
	now := time.Now().UnixMicro()

	// Lua 脚本逻辑：
	// 1. 获取当前桶内令牌数和上次刷新时间
	// 2. 根据时间差计算新生成的令牌
	// 3. 判断令牌是否足够，如果足够则扣除并更新状态
	script := redis.NewScript(1, `
		local key = KEYS[1]
		local rate = tonumber(ARGV[1])        -- 每秒生成速率
		local capacity = tonumber(ARGV[2])    -- 桶容量
		local required = tonumber(ARGV[3])    -- 需要消耗的令牌
		local now = tonumber(ARGV[4])         -- 当前时间(微秒)

		-- 获取当前状态
		local info = redis.call("HMGET", key, "tokens", "last_refill")
		local tokens = tonumber(info[1])
		local last_refill = tonumber(info[2])

		-- 如果不存在，初始化状态
		if tokens == nil then
			tokens = capacity
			last_refill = now
		end

		-- 计算时间差 (微秒) 并补充令牌
		local delta = math.max(0, now - last_refill)
		-- 生成令牌数 = 时间差(微秒) * 速率(秒) / 1,000,000
		local filled = delta * rate / 1000000
		
		-- 更新令牌数，不能超过容量
		tokens = math.min(capacity, tokens + filled)

		local allowed = 0
		if tokens >= required then
			allowed = 1
			tokens = tokens - required
			
			-- 更新 Redis 状态
			redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
			
			-- 设置过期时间，防止冷数据长期占用内存
			-- 过期时间设为填满桶所需时间的 2 倍，至少 60 秒
			local expire_time = math.ceil(capacity / rate * 2)
			if expire_time < 60 then expire_time = 60 end
			redis.call("EXPIRE", key, expire_time)
		end

		return allowed
	`)

	// 执行脚本
	res, err := redis.Int(script.Do(conn, key, rate, capacity, required, now))
	if err != nil {
		return false, err
	}

	return res == 1, nil
}
