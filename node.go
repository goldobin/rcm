package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
)

type Node struct {
	port         int
	confFilePath string
	conf         RedisNodeConf
}

func NewNode(clusterBaseDir string, port int, clusterConf ClusterConf) Node {

	baseDir := path.Join(clusterBaseDir, strconv.Itoa(port))

	return Node{
		port:         port,
		confFilePath: path.Join(baseDir, "conf", "redis.conf"),
		conf: RedisNodeConf{
			ListenHost:  clusterConf.ListenHost,
			ListenPort:  port,
			Persistence: clusterConf.Persistence,
			LogFile:     path.Join(baseDir, "var", "log", "redis.log"),
			PidFile:     path.Join(baseDir, "var", "run", "redis.pid"),
			DataDir:     path.Join(baseDir, "var", "lib", "redis"),
		},
	}
}

func (self Node) Create() {
	os.MkdirAll(path.Dir(self.confFilePath), 0750)
	os.MkdirAll(path.Dir(self.conf.LogFile), 0750)
	os.MkdirAll(path.Dir(self.conf.PidFile), 0750)
	os.MkdirAll(self.conf.DataDir, 0750)

	SaveRedisConf(self.confFilePath, &self.conf)
}

func (self Node) Start() {
	exec.Command("redis-server", self.confFilePath).Run()
}

func (self Node) Stop() {
	self.KillWithSignal("TERM")
}

func (self Node) Kill() {
	self.KillWithSignal("KILL")
}

func (self Node) KillWithSignal(signal string) {
	pid, err := self.Pid()

	if err != nil {
		panic(err)
	}

	exec.Command("kill", "-s", signal, strconv.Itoa(pid)).Run()
}

func (self Node) Cli() {
	binary, err := exec.LookPath("redis-cli")
	if err != nil {
		panic(err) // TODO: Make proper error handling
	}

	args := []string{"redis-cli", "-c", "-h", self.conf.ListenHost, "-p", strconv.Itoa(self.conf.ListenPort)}

	err = syscall.Exec(binary, args, os.Environ())

	if err != nil {
		panic(err) // TODO: Make proper error handling
	}
}

func (self Node) Pid() (pid int, err error) {
	pidBytes, err := ioutil.ReadFile(self.conf.PidFile)

	if err != nil {
		return
	}

	pid, err = strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	return
}
