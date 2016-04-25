package main

import (
	"github.com/juju/errors"
	log "github.com/ngaut/logging"
	"time"
)

func Sentinel() {
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
