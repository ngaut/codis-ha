package main

import (
	"fmt"
	//"github.com/CodisLabs/codis/pkg/utils/errors"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/docopt/docopt-go"
	//log "github.com/ngaut/logging"
	"strconv"
	"time"
)

type fnHttpCall func(objPtr interface{}, api string, method string, arg interface{}) error
type codisCheckerFactory func(addr string, defaultTimeout time.Duration) CodisChecker

var args struct {
	apiServer     string
	zookeeper     string
	quiet         bool
	slotNum       int
	fromGroupId   int
	targetGroupId int
	kill          bool
}

var (
	version  string              = "0.2.0"
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

func init() {
	log.SetLevel(log.LEVEL_INFO)
}

func main() {
	usage := `
Usage:
	codis-ha sentinel    [--server=S]
	codis-ha latency     [--server=S] [--quiet]
	codis-ha migrate     [--server=S] [--zookeeper=Z] [--kill] [--num=N] [--from=F] [--target=T]
	codis-ha version

Options:
	-s S, --server=S             Set api server address, default is "localhost:18087".
	-q, --quiet            		 Set latency output less information without slot latency.
	-z Z, --zookeeper=Z          Set zookeeper address, default is "localhost:2181".
	-k, --kill                   Kill the already running migrate tasks.
	-n N, --num=N                The number slot want to move, if the specified number bigger than slot number in from group, then move all slot in from group.
	-f F, --from=F 				 Specify the from group id ,where move the slots from.
	-t T, --target=T             Specify the target group id ,where move the slots to.
`
	d, err := docopt.Parse(usage, nil, true, "", false)
	if err != nil {
		log.Errorf("parse arguments failed: %q", err)
	}

	if args.apiServer, _ = d["--server"].(string); args.apiServer == "" {
		log.Panic("--server parameter needed")
	}

	args.quiet, _ = d["--quiet"].(bool)

	if args.zookeeper, _ = d["--zookeeper"].(string); args.zookeeper == "" {
		log.Panic("--zookeeper parameter needed")
	}

	args.kill, _ = d["--kill"].(bool)

	if d["--num"] != nil {
		args.slotNum, err = strconv.Atoi(d["--num"].(string))
		if err != nil {
			log.Panicf("parse --num failed %s", err)
		}
	}
	if d["--from"] != nil {
		args.fromGroupId, err = strconv.Atoi(d["--from"].(string))
		if err != nil {
			log.Panicf("parse --from failed %s", err)
		}
	}
	if d["--target"] != nil {
		args.targetGroupId, err = strconv.Atoi(d["--target"].(string))
		if err != nil {
			log.Panicf("parse --target failed %s", err)
		}
	}

	switch {
	case d["version"].(bool):
		fmt.Printf("Version %s\n", version)
	case d["sentinel"].(bool):
		Sentinel()
	case d["latency"].(bool):
		new(cmdLatency).Main()
	case d["migrate"].(bool):
		new(cmdMigrate).Main()
	}
}
