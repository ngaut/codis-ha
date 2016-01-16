package main

import (
	"flag"
	"github.com/juju/errors"
	log "github.com/ngaut/logging"
	"strconv"
	"time"
)

type fnHttpCall func(objPtr interface{}, api string, method string, arg interface{}) error
type aliveCheckerFactory func(addr string, defaultTimeout time.Duration) AliveChecker

var (
	apiServer   = flag.String("codis-config", "localhost:18087", "api server address")
	productName = flag.String("productName", "test", "product name, can be found in codis-proxy's config")
	logLevel    = flag.String("log-level", "info", "log level")

	callHttp fnHttpCall          = httpCall
	acf      aliveCheckerFactory = func(addr string, timeout time.Duration) AliveChecker {
		return &redisChecker{
			addr:           addr,
			defaultTimeout: timeout,
		}
	}
)

func genUrl(args ...interface{}) string {
	url := "http://"
	for _, v := range args {
		switch v.(type) {
		case string:
			url += v.(string)
		case int:
			url += strconv.Itoa(v.(int))
		default:
			log.Errorf("unsupported type %T", v)
		}
	}

	return url
}

func main() {
	flag.Parse()
	log.SetLevelByString(*logLevel)

	for {
		groups, err := GetServerGroups()
		if err != nil {
			log.Error(errors.ErrorStack(err))
			return
		}

		CheckAliveAndPromote(groups)
		CheckOfflineAndPromoteSlave(groups)
		time.Sleep(3 * time.Second)
	}
}
