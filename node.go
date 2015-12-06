package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
)

type NodeAddress struct {
	Ip   string
	Port int
}

func (self NodeAddress) String() string {
	return fmt.Sprintf("%s:%v", self.Ip, self.Port)
}

func NewNodeAddress(ip string, port int) NodeAddress {
	return NodeAddress{
		Ip:   ip,
		Port: port,
	}
}

type Node struct {
	address      NodeAddress
	confFilePath string
	conf         RedisNodeConf
	binaries     *Binaries
}

func NewNode(clusterBaseDir string, port int, clusterConf *ClusterConf, binaries *Binaries) *Node {

	baseDir := path.Join(clusterBaseDir, strconv.Itoa(port))

	return &Node{
		address:      NewNodeAddress(clusterConf.ListenIp, port),
		confFilePath: path.Join(baseDir, "conf", "redis.conf"),
		conf: RedisNodeConf{
			ListenIp:    clusterConf.ListenIp,
			ListenPort:  port,
			Persistence: clusterConf.Persistence,
			LogFile:     path.Join(baseDir, "var", "log", "redis.log"),
			PidFile:     path.Join(baseDir, "var", "run", "redis.pid"),
			DataDir:     path.Join(baseDir, "var", "lib", "redis"),
		},
		binaries: binaries,
	}
}

func (self *Node) Create() error {

	if err := os.MkdirAll(path.Dir(self.confFilePath), 0750); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(self.conf.LogFile), 0750); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(self.conf.PidFile), 0750); err != nil {
		return err
	}

	if err := os.MkdirAll(self.conf.DataDir, 0750); err != nil {
		return err
	}

	return SaveRedisConf(self.confFilePath, &self.conf)
}

func (self *Node) Start() error {
	binary := self.binaries.RedisServer()
	return exec.Command(binary, self.confFilePath).Run()
}

func (self *Node) Stop() error {
	return self.KillWithSignal("TERM")
}

func (self *Node) Kill() error {
	return self.KillWithSignal("KILL")
}

func (self *Node) KillWithSignal(signal string) error {
	pid, err := self.Pid()

	if err != nil {
		panic(err)
		// TODO: Make proper error handling
	}

	binary := self.binaries.Kill()

	return exec.Command(binary, "-s", signal, strconv.Itoa(pid)).Run()
}

func (self *Node) clientArgs(args []string) []string {
	return append(
		[]string{
			"-c",
			"-h", self.conf.ListenIp,
			"-p", strconv.Itoa(self.conf.ListenPort),
		},
		args...)
}

func (self *Node) Cli(args ...string) error {
	clientPath := self.binaries.RedisClient()
	commandArgs := append([]string{clientPath}, self.clientArgs(args)...)

	return syscall.Exec(clientPath, commandArgs, os.Environ())
}

func (self *Node) Client(args ...string) *exec.Cmd {
	binary := self.binaries.RedisClient()
	return exec.Command(binary, self.clientArgs(args)...)
}

func (self *Node) ClusterMeet(nodeAddress NodeAddress) *exec.Cmd {
	return self.Client("CLUSTER", "MEET", nodeAddress.Ip, strconv.Itoa(nodeAddress.Port))
}

func (self *Node) ClusterReplicate(id string) *exec.Cmd {
	return self.Client("CLUSTER", "REPLICATE", id)
}

func (self *Node) ClusterAddSlots(fromSlot int, toSlot int) *exec.Cmd {

	slots := make([]string, toSlot-fromSlot)

	for i := fromSlot; i < toSlot; i++ {
		slots[i-fromSlot] = strconv.Itoa(i)
	}

	args := append([]string{"CLUSTER", "ADDSLOTS"}, slots...)
	return self.Client(args...)
}

func (self *Node) ClusterSlots() *exec.Cmd {
	return self.Client("--no-raw", "CLUSTER", "SLOTS")
}

func (self *Node) ClusterNodes() *exec.Cmd {
	return self.Client("CLUSTER", "NODES")
}

func (self *Node) ClusterInfo() *exec.Cmd {
	return self.Client("CLUSTER", "INFO")
}

func (self *Node) Pid() (int, error) {
	_, statErr := os.Stat(self.conf.PidFile)
	if os.IsNotExist(statErr) {
		return -1, nil
	} else {
		pidBytes, err := ioutil.ReadFile(self.conf.PidFile)

		if err != nil {
			return -1, err
		}

		return strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	}
}

func (self *Node) Address() NodeAddress {
	return self.address
}

func (self *Node) IsUp() (result bool, err error) {
	pid, err := self.Pid()

	if err != nil {
		return
	}

	result = pid > 0

	// TODO: Check also process running and is redis-server instance
	return
}

func (self *Node) Id() (result string, err error) {
	cmd := self.ClusterNodes()

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return
	}

	err = cmd.Start()

	if err != nil {
		return
	}

	b, err := ioutil.ReadAll(stdout)

	if err != nil {
		return
	}

	err = cmd.Wait()

	if err != nil {
		return
	}

	for _, s := range strings.Split(string(b), "\n") {
		if strings.Contains(s, "myself") && len(s) > 40 {
			result = s[:40]
			return
		}
	}

	err = errors.New("Can't fetch node's id")
	return
}
