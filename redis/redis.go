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
	DefaultMaxIdle      = 10
	DefaultMaxActive    = 10
)

type Client struct {
	pool *redis.Pool
}

type Config struct {
	address  string
	auth     string
	database int64

	maxIdle        int
	maxActive      int
	idleTimeout    time.Duration
	connectTimeout time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
}

func New(cfg *Config) (*Client, error) {
	cfg.idleTimeout = DefaultIdleTimeout
	cfg.connectTimeout = DefaultConnTimeout
	cfg.readTimeout = DefaultReadTimeout
	cfg.writeTimeout = DefaultWriteTimeout
	pool := &redis.Pool{
		MaxIdle:     cfg.maxIdle,
		MaxActive:   cfg.maxActive,
		IdleTimeout: cfg.idleTimeout,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.address,
				redis.DialConnectTimeout(cfg.connectTimeout),
				redis.DialReadTimeout(cfg.readTimeout),
				redis.DialWriteTimeout(cfg.writeTimeout),
			)
			if err != nil {
				return nil, err
			}
			if cfg.auth != "" {
				_, err = c.Do("AUTH", cfg.auth)
				if err != nil {
					defer c.Close()
					return nil, err
				}
			}

			if cfg.database > 0 {
				_, err = c.Do("SELECT", cfg.database)
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

// 返回 int
func (r *Client) int(cmd string, args ...interface{}) (int, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// 返回 int64
func (r *Client) int64(cmd string, args ...interface{}) (int64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int64(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// 返回 uint64
func (r *Client) uint64(cmd string, args ...interface{}) (uint64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Uint64(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// 返回 float64
func (r *Client) float64(cmd string, args ...interface{}) (float64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Float64(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return 0, nil
	}

	return v, err
}

// 返回 string
func (r *Client) string(cmd string, args ...interface{}) (string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.String(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return "", nil
	}

	return v, err
}

// 返回 bytes
func (r *Client) bytes(cmd string, args ...interface{}) ([]byte, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Bytes(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 bool
func (r *Client) bool(cmd string, args ...interface{}) (bool, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Bool(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return false, nil
	}

	return v, err
}

// 返回 []interface{}
func (r *Client) values(cmd string, args ...interface{}) ([]interface{}, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Values(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 []float64
func (r *Client) float64s(cmd string, args ...interface{}) ([]float64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Float64s(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 []string
func (r *Client) strings(cmd string, args ...interface{}) ([]string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Strings(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 [][]byte
func (r *Client) byteSlices(cmd string, args ...interface{}) ([][]byte, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.ByteSlices(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 []int64
func (r *Client) int64s(cmd string, args ...interface{}) ([]int64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int64s(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 []int
func (r *Client) ints(cmd string, args ...interface{}) ([]int, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Ints(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 map[string]string
func (r *Client) stringMap(cmd string, args ...interface{}) (map[string]string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.StringMap(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 map[string]int
func (r *Client) intMap(cmd string, args ...interface{}) (map[string]int, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.IntMap(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 map[string]int64
func (r *Client) int64Map(cmd string, args ...interface{}) (map[string]int64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Int64Map(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}

// 返回 positions
func (r *Client) positions(cmd string, args ...interface{}) ([]*[2]float64, error) {
	conn := r.pool.Get()
	defer conn.Close()

	reply, e := conn.Do(cmd, args...)
	v, err := redis.Positions(reply, e)
	if errors.As(err, &redis.ErrNil) {
		return nil, nil
	}

	return v, err
}
