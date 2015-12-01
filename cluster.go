package main

import (
	"math/rand"
	"time"
)

const RedisSlotCount int = 16384

type Cluster struct {
	nodes []Node
}

type ClusterStats struct {
	nodesTotal int
	nodesUp    int
}

type Shard struct {
	Master   NodeAddress
	Slaves   []NodeAddress
	FromSlot int
	ToSlot   int
}

func NewCluster(baseDir string, conf ClusterConf) Cluster {

	nodes := make([]Node, len(conf.ListenPorts))

	for i, port := range conf.ListenPorts {
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

func (self Cluster) NodesCount() int {
	return len(self.nodes)
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

func (self Cluster) Cli(args []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	nodeIx := r.Intn(len(self.nodes))
	self.nodes[nodeIx].Cli(args)
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

func (self Cluster) PrepareSlotDistribution(replicas int) []Shard {
	if replicas < 0 {
		panic("Number of replicas should be greater than zero")
	}

	nodesWithDataCount := replicas + 1

	nodesCount := len(self.nodes)

	if nodesWithDataCount >= nodesCount {
		panic("Number of replicas should be less then nodes count")
	}

	mastersCount := nodesCount / nodesWithDataCount

	result := make([]Shard, mastersCount)

	slotsPerShard := RedisSlotCount / mastersCount

	for i, node := range self.nodes[:mastersCount] {
		result[i].Master = node.Address()

		fromSlot := slotsPerShard * i

		result[i].FromSlot = fromSlot
		result[i].ToSlot = fromSlot + slotsPerShard
	}

	result[mastersCount-1].ToSlot = RedisSlotCount

	for j, node := range self.nodes[mastersCount:] {
		i := j % mastersCount
		result[i].Slaves = append(result[i].Slaves, node.Address())
	}

	return result
}
