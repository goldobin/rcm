package main

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ClusterConf struct {
	ListenHost  string `yaml:"bind"`
	Ports       []int
	Persistence bool
}

func LoadClusterConf(fileName string) (r *ClusterConf, err error) {

	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	r = &ClusterConf{}

	err = yaml.Unmarshal(data, r)

	if err != nil {
		return nil, err
	}

	return r, nil
}

func SaveClusterConf(fileName string, conf *ClusterConf) (err error) {

	data, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, data, 0644)
}

type RedisNodeConf struct {
	ListenHost  string
	ListenPort  int
	Persistence bool
	DataDir     string
	PidFile     string
	LogFile     string
}

func SaveRedisConf(fileName string, conf *RedisNodeConf) (err error) {

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	w.WriteString("daemonize yes\n")
	w.WriteString("cluster-enabled yes\n")
	w.WriteString("loglevel notice\n")

	if len(conf.ListenHost) > 0 {
		fmt.Fprintf(w, "bind %s\n", conf.ListenHost)
	}

	if conf.ListenPort > 0 {
		fmt.Fprintf(w, "port %d\n", conf.ListenPort)
	}

	if len(conf.PidFile) > 0 {
		fmt.Fprintf(w, "pidfile %s\n", conf.PidFile)
	}

	if len(conf.LogFile) > 0 {
		fmt.Fprintf(w, "logfile %s\n", conf.LogFile)
	}

	if conf.Persistence {
		w.WriteString("appendonly yes\n")
		fmt.Fprintf(w, "dir %s\n", conf.DataDir)
	} else {
		w.WriteString("appendonly no\n")
		w.WriteString("save \"\"\n")
	}

	w.Flush()

	return nil
}
