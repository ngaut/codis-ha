package main

import (
	"github.com/juju/errors"
	log "github.com/ngaut/logging"
	"github.com/wandoulabs/codis/pkg/models"
	"time"
)

func GetServerGroups() ([]models.ServerGroup, error) {
	var groups []models.ServerGroup
	err := callHttp(&groups, genUrl(*apiServer, "/api/server_groups"), "GET", nil)
	return groups, err
}

func PingServer(checker AliveChecker, errCtx interface{}, errCh chan<- interface{}) {
	err := checker.CheckAlive()
	log.Debugf("check %+v, result:%v, errCtx:%+v", checker, err, errCtx)
	if err != nil {
		errCh <- errCtx
		return
	}
	errCh <- nil
}

func verifyAndUpServer(checker AliveChecker, errCtx interface{}) {
	errCh := make(chan interface{}, 100)

	go PingServer(checker, errCtx, errCh)

	s := <-errCh

	if s == nil { //alive
		handleAddServer(errCtx.(*models.Server))
	}

}

func getSlave(master *models.Server) (*models.Server, error) {
	var group models.ServerGroup
	err := callHttp(&group, genUrl(*apiServer, "/api/server_group/", master.GroupId), "GET", nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	for _, s := range group.Servers {
		if s.Type == models.SERVER_TYPE_SLAVE {
			return s, nil
		}
	}

	return nil, errors.Errorf("can not find any slave in this group: %v", group)
}

func promoteRedis(checker AliveChecker) {
	err := checker.Promote()
	log.Debugf("promote redis from slave to master %+v, result:%v, errCtx:%+v", checker, err)
	if err != nil {
		log.Errorf("do promote redis from slave to master %v failed %v", checker, errors.ErrorStack(err))
		return
	}
	return
}

func handleCrashedServer(s *models.Server) error {
	switch s.Type {
	case models.SERVER_TYPE_MASTER:
		//get slave and do promote
		slave, err := getSlave(s)
		if err != nil {
			log.Warning(errors.ErrorStack(err))
			return err
		}

		log.Infof("try promote %+v", slave)
		rc := acf(slave.Addr, 5*time.Second)
		promoteRedis(rc)
		err = callHttp(nil, genUrl(*apiServer, "/api/server_group/", slave.GroupId, "/promote"), "POST", slave)
		if err != nil {
			log.Errorf("do promote %v failed %v", slave, errors.ErrorStack(err))
			return err
		}
	case models.SERVER_TYPE_SLAVE:
		log.Errorf("slave is down: %+v", s)
	case models.SERVER_TYPE_OFFLINE:
		//no need to handle it
	default:
		log.Fatalf("unkonwn type %+v", s)
	}

	return nil
}

func handleAddServer(s *models.Server) {
	s.Type = models.SERVER_TYPE_SLAVE
	log.Infof("try reusing slave %+v", s)
	err := callHttp(nil, genUrl(*apiServer, "/api/server_group/", s.GroupId, "/addServer"), "PUT", s)
	log.Errorf("do reusing slave %v failed %v", s, errors.ErrorStack(err))
}

//ping codis-server find crashed codis-server
func CheckAliveAndPromote(groups []models.ServerGroup) ([]models.Server, error) {
	errCh := make(chan interface{}, 100)
	var serverCnt int
	for _, group := range groups { //each group
		for _, s := range group.Servers { //each server
			serverCnt++
			rc := acf(s.Addr, 5*time.Second)
			news := s
			go PingServer(rc, news, errCh)
		}
	}

	//get result
	var crashedServer []models.Server
	for i := 0; i < serverCnt; i++ {
		s := <-errCh
		if s == nil { //alive
			continue
		}

		log.Warningf("server maybe crashed %+v", s)
		crashedServer = append(crashedServer, *s.(*models.Server))

		err := handleCrashedServer(s.(*models.Server))
		if err != nil {
			return crashedServer, err
		}
	}

	return crashedServer, nil
}

//ping codis-server find node up with type offine
func CheckOfflineAndPromoteSlave(groups []models.ServerGroup) ([]models.Server, error) {
	for _, group := range groups { //each group
		for _, s := range group.Servers { //each server
			rc := acf(s.Addr, 5*time.Second)
			news := s
			if (s.Type == models.SERVER_TYPE_OFFLINE) {
				verifyAndUpServer(rc, news)
			}
		}
	}
	return nil, nil
}
