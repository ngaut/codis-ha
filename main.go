package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/juju/errors"
	log "github.com/ngaut/logging"
)

var (
	defaultConfigFile string = "./codis-ha.json"

	defaultDashboardAddr string = "127.0.0.1:18087"
	defaultProductName   string = "db_test"
	defaultLogFile       string = "./codis-ha.log"
	defaultLogLevel      string = "info"
	defaultCheckInterval int    = 5
	defaultMaxTryTimes   int    = 10
	defaultEmailAddr     string = ""
	defaultEmailPwd      string = ""
	defaultSmtpAddr      string = ""
	defaultToAddr        string = ""
	defaultSendInterval  int64  = 300
	defaultMasterSave    string = ""
	defaultSlaveSave     string = "600 1"
)

type CodisHAConf struct {
	DashboadAddr  string `json:"dashboard_addr"`
	ProductName   string `json:"product_name"`
	LogFile       string `json:"log_file"`
	LogLevel      string `json:"log_level"`
	CheckInterval int    `json:"check_interval"`
	MaxTryTimes   int    `json:"max_try_times"`
	EmailAddr     string `json:"email_addr"`
	EmailPwd      string `json:"email_pwd"`
	SmtpAddr      string `json:"smtp_addr"`
	ToAddr        string `json:"to_addr"`
	SendInterval  int64  `json:"send_interval"`
	MasterSave    string `json:"master_save"`
	SlaveSave     string `json:"slave_save"`
}

var HAConf CodisHAConf = CodisHAConf{
	DashboadAddr:  defaultDashboardAddr,
	ProductName:   defaultProductName,
	LogFile:       defaultLogFile,
	LogLevel:      defaultLogLevel,
	CheckInterval: defaultCheckInterval,
	MaxTryTimes:   defaultMaxTryTimes,
	EmailAddr:     defaultEmailAddr,
	EmailPwd:      defaultEmailPwd,
	SmtpAddr:      defaultSmtpAddr,
	ToAddr:        defaultToAddr,
	SendInterval:  defaultSendInterval,
	MasterSave:    defaultMasterSave,
	SlaveSave:     defaultSlaveSave,
}

func LoadConf(fileName string, v interface{}) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), v)
	if err != nil {
		return err
	}
	return nil
}

func PrintCodisHAConf(conf CodisHAConf) {
	fmt.Printf("dashboard_addr:%s\n", conf.DashboadAddr)
	fmt.Printf("product_name:%s\n", conf.ProductName)
	fmt.Printf("LogFile:%s\n", conf.LogFile)
	fmt.Printf("LogLevel:%s\n", conf.LogLevel)
	fmt.Printf("CheckInterval:%d\n", conf.CheckInterval)
	fmt.Printf("MaxTryTimes:%d\n", conf.MaxTryTimes)
	fmt.Printf("EmailAddr:%s\n", conf.EmailAddr)
	fmt.Printf("EmailPwd:%s\n", conf.EmailPwd)
	fmt.Printf("SmtpAddr:%s\n", conf.SmtpAddr)
	fmt.Printf("ToAddr:%s\n", conf.ToAddr)
	fmt.Printf("SendInterval:%d\n", conf.SendInterval)
	fmt.Printf("MasterSave:%d\n", conf.MasterSave)
	fmt.Printf("SlaveSave:%d\n", conf.SlaveSave)
}

type fnHttpCall func(objPtr interface{}, api string, method string, arg interface{}) error
type aliveCheckerFactory func(addr, role string, defaultTimeout time.Duration) AliveChecker

var (
	callHttp fnHttpCall          = httpCall
	acf      aliveCheckerFactory = func(addr, role string, timeout time.Duration) AliveChecker {
		return &redisChecker{
			addr:           addr,
			role:           role,
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

func Usage(progName string) {
	fmt.Printf("Usage: %s [xxx.json]\n", progName)
	os.Exit(0)
}

var JsonExample string = `
{
	"dashboard_addr":"127.0.0.1:18087",
	"product_name":"db_test",
	"log_file":"./codis-ha.log",
	"log_level":"info",
	"check_interval":5,
	"max_try_times":10,
	"email_addr":"xxx@letv.com",
	"email_pwd":"xxx",
	"smtp_addr":"mail.letv.com:25",
	"to_addr":"xxx@126.com;xxx@163.com",
	"send_interval":300,
    "master_save":"",
	"slave_save":"600 1"
}
`

func ShowJsonExample(config string) {
	fmt.Printf("%s should like this:\n%s\n", config, JsonExample)
	os.Exit(0)
}

func main() {
	var argNum int = len(os.Args)
	var confile string

	if argNum != 1 && argNum != 2 {
		Usage(os.Args[0])
	}
	if argNum == 1 {
		confile = defaultConfigFile
	} else if argNum == 2 {
		confile = os.Args[1]
	}
	err := LoadConf(confile, &HAConf)
	if err != nil {
		fmt.Printf("Load config [%s] failed,err:%s\n", confile, err.Error())
		ShowJsonExample(confile)
	}
	if len(HAConf.LogFile) > 0 {
		log.SetOutputByName(HAConf.LogFile)
	}
	if len(HAConf.LogLevel) > 0 {
		log.SetLevelByString(HAConf.LogLevel)
	}
	//PrintCodisHAConf(HAConf)
	log.Infof("program [%s] start...", os.Args[0])
	for {
		groups, err := GetServerGroups()
		if err != nil {
			log.Errorf("GetServerGroups failed,will sleep 30 seconds,err:%s", errors.ErrorStack(err))
			time.Sleep(30 * time.Second)
			continue
		}

		CheckAliveAndPromote(groups)
		CheckOfflineAndPromoteSlave(groups)
		time.Sleep(time.Duration(HAConf.CheckInterval) * time.Second)
	}
}
