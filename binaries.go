package main

import "os/exec"

type Binaries struct {
	binaries map[string]string
}

func NewBinaries() (*Binaries, error) {

	binaries := map[string]string{
		"redis-server": "",
		"redis-cli":    "",
		"kill":         "",
	}

	for command, _ := range binaries {
		binary, err := exec.LookPath(command)

		if err != nil {
			return nil, err
		}

		binaries[command] = binary
	}

	return &Binaries{binaries: binaries}, nil
}

func (self *Binaries) RedisServer(args ...string) *exec.Cmd {
	return exec.Command(self.RedisServerPath(), args...)
}

func (self *Binaries) RedisClient(args ...string) *exec.Cmd {
	return exec.Command(self.RedisClientPath(), args...)
}

func (self *Binaries) Kill(args ...string) *exec.Cmd {
	return exec.Command(self.KillPath(), args...)
}

func (self *Binaries) RedisServerPath() string {
	return self.binaries["redis-server"]
}

func (self *Binaries) RedisClientPath() string {
	return self.binaries["redis-cli"]
}

func (self *Binaries) KillPath() string {
	return self.binaries["kill"]
}
