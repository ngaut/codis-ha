package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/juju/errors"
	log "github.com/ngaut/logging"
	"github.com/wandoulabs/codis/pkg/models"
)

var (
	apiServer   = flag.String("apiserver", "localhost:18087", "api server address")
	productName = flag.String("productName", "test", "product name, can be found in codis-proxy's config")

	tr = http.DefaultTransport
)

func getApiResult(result interface{}, api string, method string, arg interface{}) error {
	client := &http.Client{Transport: tr}
	url := "http://" + *apiServer + api
	rw := &bytes.Buffer{}
	if arg != nil {
		buf, err := json.Marshal(arg)
		if err != nil {
			return errors.Trace(err)
		}
		rw.Write(buf)
	}

	req, err := http.NewRequest(method, url, rw)
	if err != nil {
		return errors.Trace(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Trace(err)
	}

	if resp.StatusCode/100 != 2 {
		return errors.Errorf("error: %d, message: %s", resp.StatusCode, string(buf))
	}

	if result != nil {
		return errors.Trace(json.Unmarshal(buf, result))
	}

	return nil
}

func GetServerGroups() ([]models.ServerGroup, error) {
	var groups []models.ServerGroup
	err := getApiResult(&groups, "/api/server_groups", "GET", nil)
	return groups, err
}

func DoPing(instance AliveChecker) error {
	return instance.CheckAlive()
}

func PingServer(s models.Server, errServerCh chan<- *models.Server) {
	rc := &redisChecker{
		addr:           s.Addr,
		defaultTimeout: 5 * time.Second,
	}
	if err := DoPing(rc); err != nil {
		errServerCh <- &s
	} else {
		errServerCh <- nil
	}
}

func getSlave(master *models.Server) (*models.Server, error) {
	var group models.ServerGroup
	err := getApiResult(&group, "/api/server_group/"+strconv.Itoa(master.GroupId), "GET", nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	for _, s := range group.Servers {
		if s.Type == models.SERVER_TYPE_SLAVE {
			return &s, nil
		}
	}

	return nil, errors.Errorf("can not find any slave in this group %v", group)
}

func handleCrashedServer(s *models.Server) {
	switch s.Type {
	case models.SERVER_TYPE_MASTER:
		//get slave and do promote
		slave, err := getSlave(s)
		if err != nil {
			log.Warning(errors.ErrorStack(err))
			return
		}

		err = getApiResult(nil, "/api/server_group/"+strconv.Itoa(slave.GroupId)+"/promote", "POST", slave)
		if err != nil {
			log.Errorf("do promote %v failed %v", slave, errors.ErrorStack(err))
			return
		}
	case models.SERVER_TYPE_SLAVE:
		log.Errorf("server is down: %+v", s)
	case models.SERVER_TYPE_OFFLINE:
		//no need to handle it
	default:
		log.Errorf("unkonwn type %+v", s)
	}
}

//ping codis-server find crashed codis-server
func PingCrashedNodes(groups []models.ServerGroup) ([]models.ServerGroup, error) {
	errServerCh := make(chan *models.Server, 100)
	var serverCnt int
	for _, group := range groups { //each group
		for _, s := range group.Servers { //each server
			serverCnt++
			go PingServer(s, errServerCh)
		}
	}

	//get result
	for i := 0; i < serverCnt; i++ {
		s := <-errServerCh
		if s == nil { //alive
			continue
		}

		log.Warningf("server maybe crashed %+v", s)
		handleCrashedServer(s)
	}

	return nil, nil
}

func main() {
	flag.Parse()
	for {
		groups, err := GetServerGroups()
		if err != nil {
			log.Error(errors.ErrorStack(err))
			return
		}

		PingCrashedNodes(groups)
		time.Sleep(3 * time.Second)
	}
}
