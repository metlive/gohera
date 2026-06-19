package redis

import (
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	DefaultIdleTimeout  = 30 * time.Second
	DefaultConnTimeout  = 2 * time.Second
	DefaultReadTimeout  = 2 * time.Second
	DefaultWriteTimeout = 2 * time.Second
)

type Client struct {
	pool *redis.Pool
}

type Config struct {
	Address  string
	Auth     string
	Database int64

	MaxIdle        int
	MaxActive      int
	IdleTimeout    time.Duration
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// New 创建一个新的 Redis 客户端
func New(cfg *Config) (*Client, error) {
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = DefaultIdleTimeout
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = DefaultConnTimeout
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = DefaultReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = DefaultWriteTimeout
	}
	pool := &redis.Pool{
		MaxIdle:     cfg.MaxIdle,
		MaxActive:   cfg.MaxActive,
		IdleTimeout: cfg.IdleTimeout,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.Address,
				redis.DialConnectTimeout(cfg.ConnectTimeout),
				redis.DialReadTimeout(cfg.ReadTimeout),
				redis.DialWriteTimeout(cfg.WriteTimeout),
			)
			if err != nil {
				return nil, err
			}
			if cfg.Auth != "" {
				_, err = c.Do("AUTH", cfg.Auth)
				if err != nil {
					defer c.Close()
					return nil, err
				}
			}

			if cfg.Database > 0 {
				_, err = c.Do("SELECT", cfg.Database)
				if err != nil {
					defer c.Close()
					return nil, err
				}
			}

			return c, err
		},
	}

	return &Client{
		pool,
	}, nil
}

// Int 返回 int
func (r *Client) Int(cmd string, args ...any) (int, error) {
	return r.int(cmd, args...)
}

// 返回 int
func (r *Client) int(cmd string, args ...any) (int, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// Int64 返回 int64
func (r *Client) Int64(cmd string, args ...any) (int64, error) {
	return r.int64(cmd, args...)
}

// 返回 int64
func (r *Client) int64(cmd string, args ...any) (int64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int64(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// Uint64 返回 uint64
func (r *Client) Uint64(cmd string, args ...any) (uint64, error) {
	return r.uint64(cmd, args...)
}

// 返回 uint64
func (r *Client) uint64(cmd string, args ...any) (uint64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Uint64(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// Float64 返回 float64
func (r *Client) Float64(cmd string, args ...any) (float64, error) {
	return r.float64(cmd, args...)
}

// 返回 float64
func (r *Client) float64(cmd string, args ...any) (float64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Float64(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// String 返回 string
func (r *Client) String(cmd string, args ...any) (string, error) {
	return r.string(cmd, args...)
}

// 返回 string
func (r *Client) string(cmd string, args ...any) (string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.String(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return "", nil
	}

	return v, err
}

// Bytes 返回 bytes
func (r *Client) Bytes(cmd string, args ...any) ([]byte, error) {
	return r.bytes(cmd, args...)
}

// 返回 bytes
func (r *Client) bytes(cmd string, args ...any) ([]byte, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Bytes(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// Bool 返回 bool
func (r *Client) Bool(cmd string, args ...any) (bool, error) {
	return r.bool(cmd, args...)
}

// 返回 bool
func (r *Client) bool(cmd string, args ...any) (bool, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Bool(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return false, nil
	}

	return v, err
}

// Values 返回 []any
func (r *Client) Values(cmd string, args ...any) ([]any, error) {
	return r.values(cmd, args...)
}

// 返回 []any
func (r *Client) values(cmd string, args ...any) ([]any, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Values(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// Float64s 返回 []float64
func (r *Client) Float64s(cmd string, args ...any) ([]float64, error) {
	return r.float64s(cmd, args...)
}

// 返回 []float64
func (r *Client) float64s(cmd string, args ...any) ([]float64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Float64s(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// Strings 返回 []string
func (r *Client) Strings(cmd string, args ...any) ([]string, error) {
	return r.strings(cmd, args...)
}

// 返回 []string
func (r *Client) strings(cmd string, args ...any) ([]string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Strings(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// ByteSlices 返回 [][]byte
func (r *Client) ByteSlices(cmd string, args ...any) ([][]byte, error) {
	return r.byteSlices(cmd, args...)
}

// 返回 [][]byte
func (r *Client) byteSlices(cmd string, args ...any) ([][]byte, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.ByteSlices(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// Int64s 返回 []int64
func (r *Client) Int64s(cmd string, args ...any) ([]int64, error) {
	return r.int64s(cmd, args...)
}

// 返回 []int64
func (r *Client) int64s(cmd string, args ...any) ([]int64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int64s(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// Ints 返回 []int
func (r *Client) Ints(cmd string, args ...any) ([]int, error) {
	return r.ints(cmd, args...)
}

// 返回 []int
func (r *Client) ints(cmd string, args ...any) ([]int, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Ints(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// StringMap 返回 map[string]string
func (r *Client) StringMap(cmd string, args ...any) (map[string]string, error) {
	return r.stringMap(cmd, args...)
}

// 返回 map[string]string
func (r *Client) stringMap(cmd string, args ...any) (map[string]string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.StringMap(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// IntMap 返回 map[string]int
func (r *Client) IntMap(cmd string, args ...any) (map[string]int, error) {
	return r.intMap(cmd, args...)
}

// 返回 map[string]int
func (r *Client) intMap(cmd string, args ...any) (map[string]int, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.IntMap(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// Int64Map 返回 map[string]int64
func (r *Client) Int64Map(cmd string, args ...any) (map[string]int64, error) {
	return r.int64Map(cmd, args...)
}

// 返回 map[string]int64
func (r *Client) int64Map(cmd string, args ...any) (map[string]int64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int64Map(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// Positions 返回 positions
func (r *Client) Positions(cmd string, args ...any) ([]*[2]float64, error) {
	return r.positions(cmd, args...)
}

// 返回 positions
func (r *Client) positions(cmd string, args ...any) ([]*[2]float64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Positions(reply, e)
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}

	return v, err
}
