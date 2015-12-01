package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const ClusterConfFileName string = "cluster.yml"

type ClusterSet struct {
	baseDir  string
	binaries *Binaries
}

func NewClusterSet(baseDir string, binaries *Binaries) (result ClusterSet, err error) {
	err = os.MkdirAll(baseDir, 0750)

	if err != nil {
		return
	}

	result = ClusterSet{
		baseDir:  baseDir,
		binaries: binaries,
	}

	return
}

func (self ClusterSet) Create(name string, conf *ClusterConf) (result Cluster, err error) {

	if self.Exists(name) {
		err = fmt.Errorf("Cluster %s already exists", name)
		return
	}

	err = os.MkdirAll(self.clusterBaseDir(name), 0750)

	if err != nil {
		return
	}

	err = SaveClusterConf(self.clusterConfFile(name), conf)

	if err != nil {
		return
	}

	result = NewCluster(self.clusterBaseDir(name), conf, self.binaries)
	result.CreateNodes()

	return
}

func (self ClusterSet) Exists(name string) bool {
	_, err := os.Stat(self.clusterBaseDir(name))
	return !os.IsNotExist(err)
}

func (self ClusterSet) Open(name string) (result Cluster, err error) {

	if !self.Exists(name) {
		err = fmt.Errorf("Cluster %s not exists", name)
		return
	}

	conf, err := LoadClusterConf(self.clusterConfFile(name))

	if err != nil {
		return
	}

	result = NewCluster(self.clusterBaseDir(name), conf, self.binaries)
	return
}

func (self ClusterSet) Remove(name string) error {
	return os.RemoveAll(self.clusterBaseDir(name))
}

func (self ClusterSet) ListNames() (result []string) {
	files, _ := ioutil.ReadDir(self.baseDir)

	result = make([]string, len(files))

	for i, f := range files {
		result[i] = f.Name()
	}
	return
}

func (self ClusterSet) clusterBaseDir(name string) string {
	return path.Join(self.baseDir, name)
}

func (self ClusterSet) clusterConfFile(name string) string {
	return path.Join(self.clusterBaseDir(name), ClusterConfFileName)
}
