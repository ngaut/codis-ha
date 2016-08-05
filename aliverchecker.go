package main

import (
	"time"
)

type CodisChecker interface {
	CheckAlive() error
	SetLatency(key string, value int) time.Duration
}
