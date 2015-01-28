package main

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	_ AliveChecker = &redisChecker{}
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
