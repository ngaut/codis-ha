package main

import (
	"strconv"
	"time"
	"fmt"
	"github.com/docopt/docopt-go"
	log "github.com/ngaut/logging"
)

type fnHttpCall func(objPtr interface{}, api string, method string, arg interface{}) error
type codisCheckerFactory func(addr string, defaultTimeout time.Duration) CodisChecker

var args struct {
	apiServer string
	logLevel  string
	quiet     bool
}

var (
	version string = "0.1.0"
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
	codis-ha latency    [--server=S]  [--logLevel=L] [--quiet]
	codis-ha version

Options:
	-s S, --server=S                 Set api server address, default is "localhost:18087".
	-l L, --logLevel=L               Set loglevel, default is "info".
	-q, --quiet            		 Set latency output less information without slot latency.	
`
	d, err := docopt.Parse(usage, nil, true, "", false)
	if err != nil {
		log.Errorf("parse arguments failed: %q", err)
	}

	args.apiServer, _ = d["--server"].(string)
	args.logLevel, _ = d["--logLevel"].(string)
	args.quiet, _ = d["--quiet"].(bool)

	log.SetLevelByString(args.logLevel)

	switch {
	case d["version"].(bool):
		fmt.Printf("Version %s\n",version)
	case d["sentinel"].(bool):
		Sentinel()
	case d["latency"].(bool):
		new(cmdLatency).Main()
	}
}
