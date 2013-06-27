package cache

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/jiorry/gos/log"
	"strconv"
	"time"
)

type RedisCache struct {
	pool        *redis.Pool
	dbindex     int
	MaxIdle     int
	MaxActive   int
	IdleTimeout int

	EnableMutex bool
}

func (this *RedisCache) Init(data map[string]string) bool {
	this.MaxIdle = 3
	this.MaxActive = 3
	this.IdleTimeout = 300
	this.EnableMutex = true

	if len(data["max_idle"]) > 0 {
		this.MaxIdle, _ = strconv.Atoi(data["max_idle"])
	}
	if len(data["max_active"]) > 0 {
		this.MaxActive, _ = strconv.Atoi(data["max_active"])
	}
	if len(data["idle_timeout"]) > 0 {
		this.IdleTimeout, _ = strconv.Atoi(data["idle_timeout"])
	}

	this.pool = &redis.Pool{
		MaxIdle:     this.MaxIdle,
		MaxActive:   this.MaxActive,
		IdleTimeout: time.Duration(int64(this.IdleTimeout)) * time.Second,
		Dial: func() (redis.Conn, error) {
			// tcp 127.0.0.1:6379
			c, err := redis.Dial(data["network"], data["connect"])
			if err != nil {
				log.App.Crit(err)
				return nil, err
			}

			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	if len(data["db"]) > 0 {
		this.dbindex, _ = strconv.Atoi(data["db"])
	}

	conn := this.conn()
	_, err := conn.Do("PING")
	if err != nil {
		fmt.Println("...", err)
		return false
	}
	defer conn.Close()
	conn.Do("SELECT", this.dbindex)
	return true
}

func (this *RedisCache) conn() redis.Conn {
	return this.pool.Get()
}

func (this *RedisCache) Delete(key string) error {
	conn := this.conn()
	defer conn.Close()
	_, err := conn.Do("DELETE", key)
	return err
}

func (this *RedisCache) Set(key string, value interface{}, expire int) error {
	var err error
	conn := this.conn()
	defer conn.Close()

	_, err = conn.Do("SET", key, value)
	if expire > 0 {
		conn.Do("EXPIRE", key, expire)
	}

	if err != nil {
		log.App.Crit(err)
		return err
	}
	return nil
}

func (this *RedisCache) Get(key string) (interface{}, error) {
	conn := this.conn()
	defer conn.Close()
	var value interface{}
	var err error

	if this.EnableMutex {
		var exists bool
		if exists, _ = this.Exists(key); exists {
			value, err = conn.Do("GET", key)
		} else {
			if exists, _ = this.Exists("MUTEX-" + key); exists {
				for i := 0; i < 10; i++ {
					time.Sleep(10 * time.Millisecond)
					if exists, _ = this.Exists(key); exists {
						value, err = conn.Do("GET", key)
						break
					} else {
						continue
					}
				}
			} else {
				this.Set("MUTEX-"+key, true, 1)
				value = nil
			}
		}
	} else {
		value, err = conn.Do("GET", key)
	}

	if err != nil {
		log.App.Crit(err)
		return nil, err
	}
	return value, nil
}

func (this *RedisCache) Exists(key string) (bool, error) {
	conn := this.conn()
	defer conn.Close()
	return redis.Bool(conn.Do("EXISTS", key))
}

func (this *RedisCache) GetString(key string) (string, error) {
	return redis.String(this.Get(key))
}

func (this *RedisCache) GetInt(key string) (int, error) {
	return redis.Int(this.Get(key))
}

func (this *RedisCache) GetInt64(key string) (int64, error) {
	return redis.Int64(this.Get(key))
}

func (this *RedisCache) GetFloat64(key string) (float64, error) {
	return redis.Float64(this.Get(key))
}
