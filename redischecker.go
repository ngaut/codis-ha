package main

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	_ CodisChecker = &redisChecker{}
)

type redisChecker struct {
	addr           string
	defaultTimeout time.Duration
}

func (r *redisChecker) ping() error {
	c, err := redis.DialTimeout("tcp", r.addr, r.defaultTimeout, r.defaultTimeout, r.defaultTimeout)
	if err != nil {
		return err
	}

	defer c.Close()
	_, err = c.Do("ping")
	return err
}

func (r *redisChecker) CheckAlive() error {
	var err error
	for i := 0; i < 2; i++ { //try a few times
		err = r.ping()
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		return nil
	}

	return err
}

func (r *redisChecker) set(key string, value int) error {
	c, err := redis.DialTimeout("tcp", r.addr, r.defaultTimeout, r.defaultTimeout, r.defaultTimeout)
	if err != nil {
		return err
	}

	defer c.Close()
	_, err = c.Do("set", key, value)
	return err
}

func (r *redisChecker) SetLatency(key string, value int) time.Duration {
	tb := time.Now()
	err := r.set(key, value)
	if err != nil {
		return r.defaultTimeout
	}
	return time.Now().Sub(tb)
}

func (r *redisChecker) del(key string) error {
	c, err := redis.DialTimeout("tcp", r.addr, r.defaultTimeout, r.defaultTimeout, r.defaultTimeout)
	if err != nil {
		return err
	}

	defer c.Close()
	_, err = c.Do("del", key)
	return err
}

func (r *redisChecker) DeleteLatency(key string, value int) time.Duration {
	tb := time.Now()
	err := r.del(key)
	if err != nil {
		return r.defaultTimeout
	}
	return time.Now().Sub(tb)
}
