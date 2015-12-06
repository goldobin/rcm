package main

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ClusterConf struct {
	ListenIp    string `yaml:"bind"`
	ListenPorts []int  `yaml:"ports"`
	Persistence bool
}

func LoadClusterConf(fileName string) (*ClusterConf, error) {
	if data, err := ioutil.ReadFile(fileName); err != nil {
		return nil, err
	} else {
		r := &ClusterConf{}

		if err := yaml.Unmarshal(data, r); err != nil {
			return nil, err
		} else {
			return r, nil
		}
	}
}

func SaveClusterConf(fileName string, conf *ClusterConf) error {
	if data, err := yaml.Marshal(&conf); err != nil {
		return err
	} else {
		return ioutil.WriteFile(fileName, data, 0644)
	}
}

type RedisNodeConf struct {
	ListenIp    string
	ListenPort  int
	Persistence bool
	DataDir     string
	PidFile     string
	LogFile     string
}

func SaveRedisConf(fileName string, conf *RedisNodeConf) error {

	var w *bufio.Writer

	if f, err := os.Create(fileName); err != nil {
		return err
	} else {
		w = bufio.NewWriter(f)
		defer f.Close()
	}

	if _, err := w.WriteString("daemonize yes\n"); err != nil {
		return err
	}

	if _, err := w.WriteString("cluster-enabled yes\n"); err != nil {
		return err
	}

	if _, err := w.WriteString("loglevel notice\n"); err != nil {
		return err
	}

	if len(conf.ListenIp) > 0 {
		if _, err := fmt.Fprintf(w, "bind %s\n", conf.ListenIp); err != nil {
			return err
		}
	}

	if conf.ListenPort > 0 {
		if _, err := fmt.Fprintf(w, "port %d\n", conf.ListenPort); err != nil {
			return err
		}
	}

	if len(conf.PidFile) > 0 {
		if _, err := fmt.Fprintf(w, "pidfile %s\n", conf.PidFile); err != nil {
			return err
		}
	}

	if len(conf.LogFile) > 0 {
		if _, err := fmt.Fprintf(w, "logfile %s\n", conf.LogFile); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "dir %s\n", conf.DataDir); err != nil {
		return err
	}

	if conf.Persistence {
		if _, err := w.WriteString("appendonly yes\n"); err != nil {
			return err
		}
	} else {
		if _, err := w.WriteString("appendonly no\n"); err != nil {
			return err
		}

		if _, err := w.WriteString("save \"\"\n"); err != nil {
			return err
		}
	}

	return w.Flush()
}
