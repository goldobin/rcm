package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

const RcmHome string = ".rcm"
const ClusterConfFileName string = "cluster.yml"

type Clusters struct {
	baseDir string
}

func NewClusters(baseDir string) (result Clusters, err error) {

	err = os.MkdirAll(baseDir, 0750)

	if err != nil {
		return
	}

	result = Clusters{
		baseDir: baseDir,
	}

	return
}

func NewClustersInHomeDir() (_ Clusters, err error) {
	usr, err := user.Current()
	if err != nil {
		return
	}
	return NewClusters(path.Join(usr.HomeDir, RcmHome))
}

func (self Clusters) New(name string, conf ClusterConf) (result Cluster, err error) {

	if self.Exists(name) {
		err = fmt.Errorf("Cluster %s already exists", name)
		return
	}

	err = os.MkdirAll(self.clusterBaseDir(name), 0750)

	if err != nil {
		return
	}

	err = SaveClusterConf(self.clusterConfFile(name), &conf)

	if err != nil {
		return
	}

	result = Cluster{
		baseDir: self.clusterBaseDir(name),
		conf:    conf,
	}

	return
}

func (self Clusters) Exists(name string) bool {
	_, err := os.Stat(self.clusterBaseDir(name))
	return !os.IsNotExist(err)
}

func (self Clusters) Open(name string) (result Cluster, err error) {

	if !self.Exists(name) {
		err = fmt.Errorf("Cluster %s not exists", name)
		return
	}

	conf, err := LoadClusterConf(self.clusterConfFile(name))

	if err != nil {
		return
	}

	result = Cluster{
		baseDir: self.clusterBaseDir(name),
		conf:    *conf,
	}

	return
}

func (self Clusters) Remove(name string) error {
	return os.RemoveAll(self.clusterBaseDir(name))
}

func (self Clusters) ListNames() (result []string) {
	files, _ := ioutil.ReadDir(self.baseDir)

	result = make([]string, len(files))

	for i, f := range files {
		result[i] = f.Name()
	}
	return
}

func (self Clusters) clusterBaseDir(name string) string {
	return path.Join(self.baseDir, name)
}

func (self Clusters) clusterConfFile(name string) string {
	return path.Join(self.clusterBaseDir(name), ClusterConfFileName)
}

type Cluster struct {
	baseDir string
	conf    ClusterConf
}
