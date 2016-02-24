package main

import (
	"github.com/juju/errors"
	log "github.com/ngaut/logging"
	"github.com/wandoulabs/codis/pkg/models"
	//"../codis/pkg/models"
	"fmt"
	"strings"
	"time"
)

type EmailTime struct {
	warntime  int64
	errortime int64
}

var EmailSendTime map[string]*EmailTime = make(map[string]*EmailTime)

func GetServerGroups() ([]models.ServerGroup, error) {
	var groups []models.ServerGroup
	err := callHttp(&groups, genUrl(HAConf.DashboadAddr, "/api/server_groups"), "GET", nil)
	return groups, err
}

func PingServer(checker AliveChecker, server interface{}, errCh chan<- interface{}) {
	err := checker.CheckAlive()
	if err != nil {
		s := err.Error()
		if strings.Contains(s, "LOADING") {
			log.Infof("server [%+v] is LOADING... ", server)
			err = nil
		} else {
			log.Warningf("server [%+v] may be down,check failed,err:%s", server, err.Error())
		}
	} else {
		log.Debugf("server [%+v] is alive", server)
	}
	if err != nil {
		errCh <- server
		return
	}
	errCh <- nil
}

func verifyAndUpServer(checker AliveChecker, errCtx interface{}) {
	errCh := make(chan interface{}, 4)

	go PingServer(checker, errCtx, errCh)

	s := <-errCh

	if s == nil { //alive
		handleAddServer(errCtx.(*models.Server))
	}

}

func getSlave(master *models.Server) (*models.Server, error) {
	var group models.ServerGroup
	err := callHttp(&group, genUrl(HAConf.DashboadAddr, "/api/server_group/", master.GroupId), "GET", nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	for _, s := range group.Servers {
		if s.Type == models.SERVER_TYPE_SLAVE {
			// check slave is alive or not
			rc := acf(s.Addr, 3*time.Second)
			err = rc.CheckAlive()
			if err != nil {
				log.Warningf("master [%s] crashed,its slave [%s] crashed too,find next.", master.Addr, s.Addr)
				continue
			}
			return s, nil
		}
	}

	return nil, errors.Errorf("can not find any slave in this group: %v", group)
}

func handleCrashedServer(s *models.Server) error {
	switch s.Type {
	case models.SERVER_TYPE_MASTER:
		// find one valid slave and do promote
		slave, err := getSlave(s)
		if err != nil {
			log.Warningf("get slave of master [%s] failed,err:%s", s.Addr, errors.ErrorStack(err))
			// should send email
			tt := time.Now()
			emailtime := EmailSendTime[s.Addr]
			if emailtime == nil {
				emailtime = &EmailTime{}
				EmailSendTime[s.Addr] = emailtime
			}
			if tt.Unix()-emailtime.errortime > HAConf.SendInterval {
				subject := fmt.Sprintf("紧急告警-主从切换失败(%04d-%02d-%02d %02d:%02d:%02d)", tt.Year(), tt.Month(),
					tt.Day(), tt.Hour(), tt.Minute(), tt.Second())
				data := fmt.Sprintf("group [%d],master [%s] crashed, but not find any slave,err:%s", s.GroupId, s.Addr, err.Error())
				err = SendSmtpEmail(HAConf.EmailAddr, HAConf.EmailPwd, HAConf.SmtpAddr, HAConf.ToAddr, subject, data, "text")
				if err != nil {
					log.Warningf("send mail[%s] to [%s] failed, content [%s], err:%s", subject, HAConf.ToAddr, data, err.Error())
				} else {
					log.Warningf("send mail[%s] to [%s] success, content [%s]", subject, HAConf.ToAddr, data)
					emailtime.errortime = tt.Unix()
				}
			}
			return err
		}

		log.Infof("master [%s] crashed,try promote slave [%s]", s.Addr, slave.Addr)
		// promote slave
		err = callHttp(nil, genUrl(HAConf.DashboadAddr, "/api/server_group/", slave.GroupId, "/promote"), "POST", slave)
		if err != nil {
			log.Errorf("master [%s],do promote slave [%s] failed,error tarce: %v", s.Addr, slave.Addr, errors.ErrorStack(err))
			// should send email
			tt := time.Now()
			emailtime := EmailSendTime[s.Addr]
			if emailtime == nil {
				emailtime = &EmailTime{}
				EmailSendTime[s.Addr] = emailtime
			}
			if tt.Unix()-emailtime.errortime > HAConf.SendInterval {
				subject := fmt.Sprintf("紧急告警-主从切换失败(%04d-%02d-%02d %02d:%02d:%02d)", tt.Year(), tt.Month(),
					tt.Day(), tt.Hour(), tt.Minute(), tt.Second())
				data := fmt.Sprintf("group [%d],master [%s] crashed, promote slave [%s] failed,err:%s", s.GroupId, s.Addr, slave.Addr, err.Error())
				err = SendSmtpEmail(HAConf.EmailAddr, HAConf.EmailPwd, HAConf.SmtpAddr, HAConf.ToAddr, subject, data, "text")
				if err != nil {
					log.Warningf("send mail[%s] to [%s] failed, content [%s], err:%s", subject, HAConf.ToAddr, data, err.Error())
				} else {
					log.Warningf("send mail[%s] to [%s] success, content [%s]", subject, HAConf.ToAddr, data)
					emailtime.errortime = tt.Unix()
				}
			}
			return err
		} else {
			// should send email
			log.Infof("master [%s] crashed,promote slave [%s] success", s.Addr, slave.Addr)
			tt := time.Now()
			subject := fmt.Sprintf("通知-主从切换成功(%04d-%02d-%02d %02d:%02d:%02d)", tt.Year(), tt.Month(),
				tt.Day(), tt.Hour(), tt.Minute(), tt.Second())
			data := fmt.Sprintf("group [%d],master [%s] crashed, promote slave [%s] success.", s.GroupId, s.Addr, slave.Addr)
			err = SendSmtpEmail(HAConf.EmailAddr, HAConf.EmailPwd, HAConf.SmtpAddr, HAConf.ToAddr, subject, data, "text")
			if err != nil {
				log.Infof("send mail[%s] to [%s] failed, content [%s],err:%s", subject, HAConf.ToAddr, data, err.Error())
			} else {
				log.Infof("send mail[%s] to [%s] success, content [%s]", subject, HAConf.ToAddr, data)
			}
		}
	case models.SERVER_TYPE_SLAVE:
		log.Errorf("group [%d], type [slave], addr [%s] is down, not handle now", s.GroupId, s.Addr)
	case models.SERVER_TYPE_OFFLINE:
		//no need to handle it
	default:
		log.Fatalf("unkonwn type %+v", s)
	}

	return nil
}

func handleAddServer(s *models.Server) {
	s.Type = models.SERVER_TYPE_SLAVE
	log.Infof("try reusing offline to slave [%+v]", s)
	err := callHttp(nil, genUrl(HAConf.DashboadAddr, "/api/server_group/", s.GroupId, "/addServer"), "PUT", s)
	if err != nil {
		log.Errorf("do reusing slave [%+v] failed,trace:%s", s, errors.ErrorStack(err))
	} else {
		log.Infof("do reusing slave [%+v] success.", s)
	}
}

//ping codis-server find crashed codis-server
func CheckAliveAndPromote(groups []models.ServerGroup) ([]models.Server, error) {
	errCh := make(chan interface{}, 128)
	var serverCnt int
	for _, group := range groups { //each group
		for _, s := range group.Servers { //each server
			serverCnt++
			rc := acf(s.Addr, 3*time.Second)
			server := s
			log.Debugf("ping group[%d],type[%s],server[%s]", s.GroupId, s.Type, s.Addr)
			go PingServer(rc, server, errCh)
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
		str := fmt.Sprintf("%v", s)
		// send mail when server crashed
		if strings.Contains(str, models.SERVER_TYPE_OFFLINE) == false {
			tt := time.Now()
			saddr := s.(*models.Server).Addr
			emailtime := EmailSendTime[saddr]
			if emailtime == nil {
				emailtime = &EmailTime{}
				EmailSendTime[saddr] = emailtime
			}
			if tt.Unix()-emailtime.warntime > HAConf.SendInterval {
				subject := fmt.Sprintf("警告-redis故障(%04d-%02d-%02d %02d:%02d:%02d)", tt.Year(), tt.Month(),
					tt.Day(), tt.Hour(), tt.Minute(), tt.Second())
				data := fmt.Sprintf("server [%v] may be crashed.", s)
				err := SendSmtpEmail(HAConf.EmailAddr, HAConf.EmailPwd, HAConf.SmtpAddr, HAConf.ToAddr, subject, data, "text")
				if err != nil {
					log.Warningf("send mail[%s] to [%s] failed, content [%s], err:%s", subject, HAConf.ToAddr, data, err.Error())
				} else {
					log.Warningf("send mail[%s] to [%s] success, content [%s]", subject, HAConf.ToAddr, data)
					emailtime.warntime = tt.Unix()
				}
			}
		}

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
			if s.Type == models.SERVER_TYPE_OFFLINE {
				verifyAndUpServer(rc, news)
			}
		}
	}
	return nil, nil
}
