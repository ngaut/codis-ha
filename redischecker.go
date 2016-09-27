package main

import (
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	log "github.com/ngaut/logging"
	"github.com/wlibo666/codis/pkg/models"
)

var (
	_ AliveChecker = &redisChecker{}
)

type redisChecker struct {
	addr           string
	role           string
	defaultTimeout time.Duration
}

func (r *redisChecker) ping() error {
	c, err := redis.DialTimeout("tcp", r.addr, r.defaultTimeout, r.defaultTimeout, r.defaultTimeout)
	if err != nil {
		return err
	}

	defer c.Close()
	_, err = c.Do("ping")
	if err != nil {
		return err
	}
	confData := strings.Trim(HAConf.MasterSave, " ")
	if r.role == models.SERVER_TYPE_MASTER {
		_, err = c.Do("config", "set", "save", confData)
		if err != nil {
			log.Warningf("set config save [%s] for master [%s] failed,err:%s", confData, r.addr, err.Error())
		}
	}
	confData = strings.Trim(HAConf.SlaveSave, " ")
	if r.role == models.SERVER_TYPE_SLAVE {
		_, err = c.Do("config", "set", "save", confData)
		if err != nil {
			log.Warningf("set config save [%s] for slave [%s] failed,err:%s", confData, r.addr, err.Error())
		}
	}
	return err
}

func (r *redisChecker) CheckAlive() error {
	var err error
	for i := 0; i < HAConf.MaxTryTimes; i++ { //try a few times
		err = r.ping()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		return nil
	}
	return err
}
