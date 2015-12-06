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

func (self *Binaries) RedisServer() string {
	return self.binaries["redis-server"]
}

func (self *Binaries) RedisClient() string {
	return self.binaries["redis-cli"]
}

func (self *Binaries) Kill() string {
	return self.binaries["kill"]
}
