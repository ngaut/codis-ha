package main

import (
	"encoding/json"
	"fmt"
	"github.com/CodisLabs/codis/pkg/models"
	"github.com/CodisLabs/codis/pkg/utils"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/wandoulabs/zkhelper"
	"time"
)

type MigrateTaskInfo struct {
	SlotId     int    `json:"slot_id"`
	NewGroupId int    `json:"new_group"`
	Delay      int    `json:"delay"`
	CreateAt   string `json:"create_at"`
	Percent    int    `json:"percent"`
	Status     string `json:"status"`
	Id         string `json:"-"`
}

type Tasks []MigrateTaskInfo

type migrateTaskForm struct {
	From  int `json:"from"`
	To    int `json:"to"`
	Group int `json:"new_group"`
	Delay int `json:"delay"`
}

type cmdMigrate struct {
	productName  string
	safeZkConn   zkhelper.Conn
	unsafeZkConn zkhelper.Conn
	fromGroup    models.ServerGroup
	targetGroup  models.ServerGroup
}

func (cmd *cmdMigrate) Main() {
	cmd.initProductName()
	cmd.initZkConn()
	if args.kill {
		log.Info("Kill Migrating slot")
		cmd.deleteMigrateTasks()
		cmd.setSlotsOnline()
		return
	}
	cmd.initServerGroup()
	cmd.verifyBeforeMigrateTasks()

	cmd.migrateSlots(args.slotNum)
}

func (cmd *cmdMigrate) initProductName() {
	groups, err := GetServerGroups()
	if err != nil {
		log.Panic(err)
	}
	for _, g := range groups {
		cmd.productName = g.ProductName
		break
	}
}

func (cmd *cmdMigrate) verifyBeforeMigrateTasks() {
	for i := 0; i < 1000; i++ {
		cmd.checkSlotsOnline()
		tasks, err := GetMigrateTasks()
		if err != nil {
			log.Panic("get migrate status error")
		}
		if len(tasks) != 0 {
			log.Warn("There is running migrate tasks, waiting!!")
			time.Sleep(time.Second * 10)
		} else {
			return
		}
		continue
	}
	log.Panic("Wait rather long time for already running migrate tasks,quit")
}

func (cmd *cmdMigrate) checkSlotsOnline() {
	slots, err := GetSlotList()
	if err != nil {
		log.Panicf("Get Slots list err %s", err)
	}
	for _, s := range slots {
		if s.State.Status != "online" {
			log.Warnf("Slot %d status is %s", s.Id, s.State.Status)
		}
	}
}

func (cmd *cmdMigrate) initZkConn() {
	zkBuilder := utils.NewConnBuilder(cmd.newZkConn)
	cmd.safeZkConn = zkBuilder.GetSafeConn()
	cmd.unsafeZkConn = zkBuilder.GetUnsafeConn()
}

func (cmd *cmdMigrate) newZkConn() (zkhelper.Conn, error) {
	return zkhelper.ConnectToZk(args.zookeeper, 30)
}

func (cmd *cmdMigrate) initServerGroup() {
	groups, err := GetServerGroups()
	var fromFind = false
	var targetFind = false
	if args.fromGroupId < 0 || args.targetGroupId < 0 {
		log.Panicf("Both fromGroupId %d and targetGroupId should not small than 0", args.fromGroupId, args.targetGroupId)
	}
	if err != nil {
		log.Panicf("Get servergroup err %s", err)
	}
	for _, g := range groups {
		if args.fromGroupId == g.Id {
			cmd.productName = g.ProductName
			cmd.fromGroup = g
			fromFind = true
		}
		if args.targetGroupId == g.Id {
			cmd.targetGroup = g
			targetFind = true
		}
	}
	if !fromFind {
		log.Panicf("from groupId not found %d", args.fromGroupId)
	}
	if !targetFind {
		log.Panicf("target groupId not found %d", args.targetGroupId)
	}
}

func (cmd *cmdMigrate) migrateSlots(num int) {
	if num <= 0 {
		log.Panicf("slot num%d should bigger than 0 ", num)
		return
	}
	slots, err := GetSlotList()
	if err != nil {
		log.Panicf("Get Slots list err %s", err)
	}
	count := num
	for _, s := range slots {
		cmd.verifyBeforeMigrateTasks()
		if s.GroupId == cmd.fromGroup.Id {
			log.Infof("Migrate slot %d from group %d to group %d begin:", s.Id, cmd.fromGroup.Id, cmd.targetGroup.Id)
			cmd.runSlotMigrate(s.Id, s.Id, cmd.targetGroup.Id, 3)
			count--
		}
		if count <= 0 {
			log.Infof("Migrate %d slot from group %d to group %d done!", num, cmd.fromGroup.Id, cmd.targetGroup.Id)
			return
		}
	}
}

func (cmd *cmdMigrate) runSlotMigrate(fromSlotId int, toSlotId int, newGroupId int, delay int) error {
	migrateInfo := &migrateTaskForm{
		From:  fromSlotId,
		To:    toSlotId,
		Group: newGroupId,
		Delay: delay,
	}
	var v interface{}
	err := callApi(METHOD_POST, "/api/migrate", migrateInfo, &v)
	if err != nil {
		return err
	}
	fmt.Println(jsonify(v))
	return nil
}

func (cmd *cmdMigrate) tasks() []MigrateTaskInfo {
	res := Tasks{}
	tasks, _, _ := cmd.safeZkConn.Children(getMigrateTasksPath(cmd.productName))
	for _, id := range tasks {
		data, _, _ := cmd.safeZkConn.Get(getMigrateTasksPath(cmd.productName) + "/" + id)
		info := new(MigrateTaskInfo)
		json.Unmarshal(data, info)
		info.Id = id
		res = append(res, *info)
	}
	return res
}

func (cmd *cmdMigrate) deleteMigrateTasks() {
	tasks, _, _ := cmd.safeZkConn.Children(getMigrateTasksPath(cmd.productName))
	for _, id := range tasks {
		log.Warnf("Delete Migrate tasks %s", id)
		cmd.safeZkConn.Delete(getMigrateTasksPath(cmd.productName)+"/"+id, -1)
	}
}

func (cmd *cmdMigrate) setSlotsOnline() {
	slots, err := GetSlotList()
	if err != nil {
		log.Panicf("Get Slots list err %s", err)
	}
	for _, s := range slots {
		if s.State.Status != models.SLOT_STATUS_ONLINE {
			log.Warnf("Before set migrate slot %d from group %d to group %d status %s online.", s.Id, s.State.MigrateStatus.From, s.State.MigrateStatus.To, s.State.Status)
			zkPath := models.GetSlotPath(s.ProductName, s.Id)
			s.State.Status = models.SLOT_STATUS_ONLINE
			s.State.MigrateStatus.From = models.INVALID_ID
			s.State.MigrateStatus.To = models.INVALID_ID
			data, err := json.Marshal(s)
			if err != nil {
				log.Panic(err)
			}
			_, err = zkhelper.CreateOrUpdate(cmd.safeZkConn, zkPath, string(data), 0, zkhelper.DefaultFileACLs(), true)
			if err != nil {
				log.Panic(err)
			}
			log.Warnf("Set slot %d online done!!", s.Id)
		}
	}
}

func getMigrateTasksPath(product string) string {
	return fmt.Sprintf("/zk/codis/db_%s/migrate_tasks", product)
}
