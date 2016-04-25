package main

import (
	"strconv"
	"time"

	"github.com/docopt/docopt-go"
	log "github.com/ngaut/logging"
)

type fnHttpCall func(objPtr interface{}, api string, method string, arg interface{}) error
type codisCheckerFactory func(addr string, defaultTimeout time.Duration) CodisChecker

var args struct {
	apiServer string
	logLevel  string
}

var (
	callHttp fnHttpCall          = httpCall
	acf      codisCheckerFactory = func(addr string, timeout time.Duration) CodisChecker {
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
	usage := `
Usage:
	codis-ha sentinel   [--server=S]  [--logLevel=L]
	codis-ha latency  	[--server=S]  [--logLevel=L]

Options:
	-s S, --server=S                 Set api server address, default is "localhost:18087".
	-l L, --logLevel=L               Set loglevel, default is "info".
`
	d, err := docopt.Parse(usage, nil, true, "", false)
	if err != nil {
		log.Errorf("parse arguments failed: %q", err)
	}

	args.apiServer, _ = d["--server"].(string)
	args.logLevel, _ = d["--logLevel"].(string)

	log.SetLevelByString(args.logLevel)

	switch {
	case d["sentinel"].(bool):
		Sentinel()
	case d["latency"].(bool):
		new(cmdLatency).Main()
	}
}
