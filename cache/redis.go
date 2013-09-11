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

func (this *RedisCache) Delete(bkey []byte) error {
	conn := this.conn()
	defer conn.Close()
	_, err := conn.Do("DEL", stringKey(bkey))
	return err
}

func (this *RedisCache) Set(bkey []byte, value interface{}, expire int) error {
	var err error
	conn := this.conn()
	defer conn.Close()
	key := stringKey(bkey)
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

func (this *RedisCache) Get(bkey []byte) (interface{}, error) {
	conn := this.conn()
	defer conn.Close()
	var value interface{}
	var err error

	key := stringKey(bkey)
	if this.EnableMutex {
		var exists bool
		if exists, _ = this.Exists(bkey); exists {
			value, err = conn.Do("GET", key)
		} else {
			mxPre := []byte("MUTEX-")
			if exists, _ = this.Exists(append(mxPre, bkey...)); exists {
				for i := 0; i < 10; i++ {
					time.Sleep(10 * time.Millisecond)
					if exists, _ = this.Exists(bkey); exists {
						value, err = conn.Do("GET", key)
						break
					} else {
						continue
					}
				}
			} else {
				this.Set(append(mxPre, bkey...), true, 1)
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

func (this *RedisCache) Exists(bkey []byte) (bool, error) {
	conn := this.conn()
	defer conn.Close()
	return redis.Bool(conn.Do("EXISTS", stringKey(bkey)))
}

func (this *RedisCache) GetString(bkey []byte) (string, error) {
	return redis.String(this.Get(bkey))
}

func (this *RedisCache) GetInt(bkey []byte) (int, error) {
	return redis.Int(this.Get(bkey))
}

func (this *RedisCache) GetInt64(bkey []byte) (int64, error) {
	return redis.Int64(this.Get(bkey))
}

func (this *RedisCache) GetFloat64(bkey []byte) (float64, error) {
	return redis.Float64(this.Get(bkey))
}

func stringKey(bkey []byte) string {
	return string(bkey)
}
