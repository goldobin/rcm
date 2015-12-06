package main

import (
	"errors"
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

func NewClusterSet(baseDir string, binaries *Binaries) (*ClusterSet, error) {
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, err
	}

	return &ClusterSet{
			baseDir:  baseDir,
			binaries: binaries,
		},
		nil
}

func (self *ClusterSet) Create(name string, conf *ClusterConf) (*Cluster, error) {

	if self.Exists(name) {
		return nil, errors.New(fmt.Sprintf("Cluster %s already exists", name))
	}

	if err := os.MkdirAll(self.clusterBaseDir(name), 0750); err != nil {
		return nil, err
	}

	if err := SaveClusterConf(self.clusterConfFile(name), conf); err != nil {
		return nil, err
	}

	result := NewCluster(self.clusterBaseDir(name), conf, self.binaries)
	result.CreateNodes()

	return result, nil
}

func (self *ClusterSet) Exists(name string) bool {
	_, err := os.Stat(self.clusterBaseDir(name))
	return !os.IsNotExist(err)
}

func (self *ClusterSet) Open(name string) (*Cluster, error) {

	if !self.Exists(name) {
		return nil, errors.New(fmt.Sprintf("Cluster %s not exists", name))
	}

	if conf, err := LoadClusterConf(self.clusterConfFile(name)); err != nil {
		return nil, err
	} else {
		return NewCluster(self.clusterBaseDir(name), conf, self.binaries), nil
	}
}

func (self *ClusterSet) Remove(name string) error {
	return os.RemoveAll(self.clusterBaseDir(name))
}

func (self *ClusterSet) ListNames() ([]string, error) {
	if files, err := ioutil.ReadDir(self.baseDir); err != nil {
		return nil, err
	} else {
		result := make([]string, len(files))

		for i, f := range files {
			result[i] = f.Name()
		}
		return result, nil
	}
}

func (self *ClusterSet) clusterBaseDir(name string) string {
	return path.Join(self.baseDir, name)
}

func (self *ClusterSet) clusterConfFile(name string) string {
	return path.Join(self.clusterBaseDir(name), ClusterConfFileName)
}
