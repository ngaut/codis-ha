package main

import (
	"encoding/json"
	"fmt"

	"github.com/alicebob/miniredis"
	//"github.com/wandoulabs/codis/pkg/models"
	"../codis/pkg/models"
	"testing"
	"time"
)

const GROUP_ID = 1

var (
	redisServer, _ = miniredis.Run()
	groups1        = []models.ServerGroup{
		models.ServerGroup{
			Servers: []*models.Server{
				&models.Server{GroupId: GROUP_ID, Type: models.SERVER_TYPE_MASTER, Addr: "localhost:xxx"},
				&models.Server{GroupId: GROUP_ID, Type: models.SERVER_TYPE_SLAVE, Addr: redisServer.Addr()},
				&models.Server{GroupId: GROUP_ID, Type: models.SERVER_TYPE_SLAVE, Addr: "xx"},
				&models.Server{GroupId: GROUP_ID, Type: models.SERVER_TYPE_OFFLINE, Addr: "xx"},
			},
		},
	}
)

func TestGetServerGroups(t *testing.T) {
	callHttp = func(objPtr interface{}, url string, method string, arg interface{}) error {
		buf, _ := json.Marshal(groups1)
		json.Unmarshal(buf, objPtr)
		return nil
	}

	servergroups, err := GetServerGroups()
	if err != nil {
		t.Error(err)
	}

	if len(servergroups) == 0 {
		t.Error("empty server groups")
	}

	callHttp = func(objPtr interface{}, url string, method string, arg interface{}) error {
		return fmt.Errorf("mock return error")
	}

	if _, err := GetServerGroups(); err == nil {
		t.Error("should be error")
	}
}

func TestPingServer(t *testing.T) {
	rc := &redisChecker{
		defaultTimeout: 1 * time.Second,
	}

	errCh := make(chan interface{})
	go PingServer(rc, "context", errCh)
	if str := <-errCh; str.(string) != "context" {
		t.Error("should be error")
	}

	redis, _ := miniredis.Run()
	defer redis.Close()
	rc.addr = redis.Addr()
	go PingServer(rc, "context", errCh)
	if obj := <-errCh; obj != nil {
		t.Error("should be error")
	}
}

func TestGetSlave(t *testing.T) {
	callHttp = func(objPtr interface{}, url string, method string, arg interface{}) error {
		buf, _ := json.Marshal(groups1[0])
		json.Unmarshal(buf, objPtr)
		return nil
	}

	s, err := getSlave(&models.Server{})
	if err != nil {
		t.Error(err)
	}

	if s.Type != models.SERVER_TYPE_SLAVE {
		t.Error("should be slave")
	}

	callHttp = func(objPtr interface{}, url string, method string, arg interface{}) error {
		return fmt.Errorf("mock return error")
	}

	if _, err := getSlave(&models.Server{}); err == nil {
		t.Error("should be error")
	}
}

func TestCheckAliveAndPromote(t *testing.T) {
	//test promote with slave
	groups := groups1
	callHttp = func(objPtr interface{}, url string, method string, arg interface{}) error {
		if url == genUrl(HAConf.DashboadAddr, "/api/server_group/", GROUP_ID) {
			group := groups[0]
			buf, _ := json.Marshal(group)
			json.Unmarshal(buf, objPtr)
			return nil
		}

		return nil
	}

	_, err := CheckAliveAndPromote(groups)
	if err != nil {
		t.Error(err)
	}

	//test no slave
	groups = []models.ServerGroup{
		models.ServerGroup{
			Servers: []*models.Server{
				&models.Server{GroupId: GROUP_ID, Type: models.SERVER_TYPE_MASTER, Addr: "dead master"},
			},
		},
	}

	_, err = CheckAliveAndPromote(groups)
	if err == nil {
		t.Error("should be error")
	}

	//test have slave but promote error
	groups = groups1
	callHttp = func(objPtr interface{}, url string, method string, arg interface{}) error {
		if url == genUrl(HAConf.DashboadAddr, "/api/server_group/", GROUP_ID) {
			fmt.Println(url)
			group := groups[0]
			buf, _ := json.Marshal(group)
			json.Unmarshal(buf, objPtr)
			return nil
		}

		return fmt.Errorf("mock error")
	}

	_, err = CheckAliveAndPromote(groups)
	if err == nil {
		t.Error("should be error")
	}
}
