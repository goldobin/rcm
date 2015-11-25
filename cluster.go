package main

import (
	"math/rand"
	"time"
)

type Cluster struct {
	nodes []Node
}

func NewCluster(baseDir string, conf ClusterConf) Cluster {

	nodes := make([]Node, len(conf.Ports))

	for i, port := range conf.Ports {
		nodes[i] = NewNode(baseDir, port, conf)
	}

	return Cluster{
		nodes: nodes,
	}
}

func (self Cluster) CreateNodes() {
	for _, node := range self.nodes {
		node.Create()
	}
}

func (self Cluster) Start() {
	for _, node := range self.nodes {
		node.Start()
	}
}

func (self Cluster) Stop() {
	for _, node := range self.nodes {
		node.Stop()
	}
}

func (self Cluster) Kill() {
	for _, node := range self.nodes {
		node.Stop()
	}
}

func (self Cluster) Cli() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	nodeIx := r.Intn(len(self.nodes))
	self.nodes[nodeIx].Cli()
}
