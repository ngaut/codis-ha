package main

import (
	"github.com/alicebob/miniredis"
	"testing"
	"time"
)

func TestRedisChecker(t *testing.T) {
	r, _ := miniredis.Run()
	defer r.Close()
	addr := r.Addr()
	rc := &redisChecker{
		addr:           addr,
		defaultTimeout: 5 * time.Second,
	}

	err := rc.CheckAlive()
	if err != nil {
		t.Error(err)
	}

	//test bad address
	rc.addr = "xxx"
	err = rc.CheckAlive()
	if err == nil {
		t.Error("should be error")
	}
}
