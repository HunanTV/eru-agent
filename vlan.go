package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"./common"
	"./logs"
	"github.com/CMGS/consistent"
	"github.com/keimoon/gore"
)

type VLanSetter struct {
	Devices *consistent.Consistent
}

func NewVLanSetter() *VLanSetter {
	v := &VLanSetter{}
	v.Devices = consistent.New()
	for _, device := range config.VLan.Physical {
		v.Devices.Add(device)
	}
	return v
}

func (self *VLanSetter) Watcher() {
	conn, err := common.Rds.Acquire()
	if err != nil || conn == nil {
		logs.Assert(err, "Get redis conn")
	}
	defer common.Rds.Release(conn)

	subs := gore.NewSubscriptions(conn)
	defer subs.Close()
	subKey := fmt.Sprintf("eru:agent:%s:vlan", config.HostName)
	logs.Debug("Watch VLan Config", subKey)
	subs.Subscribe(subKey)

	for message := range subs.Message() {
		if message == nil {
			logs.Info("VLan Watcher Shutdown")
			break
		}
		command := string(message.Message)
		logs.Debug("Add new VLan", command)
		parser := strings.Split(command, "|")
		taskID, containerID, ident := parser[0], parser[1], parser[2]
		feedKey := fmt.Sprintf("eru:agent:%s:feedback", taskID)
		for _, seq := range parser[3:] {
			self.addVLan(feedKey, seq, ident, containerID)
		}
	}
}

func (self *VLanSetter) addVLan(feedKey, seq, ident, containerID string) {
	conn, err := common.Rds.Acquire()
	if err != nil || conn == nil {
		logs.Info(err, "Get redis conn")
		return
	}
	defer common.Rds.Release(conn)

	// Add macvlan device
	device, _ := self.Devices.Get(ident, 0)
	vethName := fmt.Sprintf("%s%s.%s", common.VLAN_PREFIX, ident, seq)
	logs.Info("Add new VLan to", vethName, containerID)
	cmd := exec.Command("ip", "link", "add", vethName, "link", device, "type", "macvlan", "mode", "bridge")
	if err := cmd.Run(); err != nil {
		gore.NewCommand("LPUSH", feedKey, fmt.Sprintf("0||")).Run(conn)
		logs.Info("Create macvlan device failed", err)
		return
	}
	container, err := common.Docker.InspectContainer(containerID)
	if err != nil {
		gore.NewCommand("LPUSH", feedKey, fmt.Sprintf("0||")).Run(conn)
		logs.Info("VLanSetter inspect docker failed", err)
		self.delVLan(vethName)
		return
	}
	cmd = exec.Command("ip", "link", "set", "netns", strconv.Itoa(container.State.Pid), vethName)
	if err := cmd.Run(); err != nil {
		gore.NewCommand("LPUSH", feedKey, fmt.Sprintf("0||")).Run(conn)
		logs.Info("Set macvlan device into container failed", err)
		self.delVLan(vethName)
		return
	}
	gore.NewCommand("LPUSH", feedKey, fmt.Sprintf("1|%s|%s", containerID, vethName)).Run(conn)
	logs.Info("Add VLAN device success", containerID, ident)
}

func (self *VLanSetter) delVLan(vethName string) {
	cmd := exec.Command("ip", "link", "del", vethName)
	cmd.Run()
}