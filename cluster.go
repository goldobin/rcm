package main

import (
	"math/rand"
	"time"
)

type Cluster struct {
	nodes []Node
}

type ClusterStats struct {
	nodesTotal int
	nodesUp    int
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

func (self Cluster) Nodes() []Node {
	result := make([]Node, len(self.nodes))
	copy(result, self.nodes)

	return result
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

func (self Cluster) Stats() (*ClusterStats, error) {

	result := ClusterStats{}
	result.nodesTotal = len(self.nodes)

	for _, node := range self.nodes {

		nodeIsUp, err := node.IsUp()

		if err != nil {
			return nil, err
		}

		if nodeIsUp {
			result.nodesUp += 1
		}
	}

	return &result, nil
}
