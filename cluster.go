package main

import (
	"math/rand"
	"sort"
	"time"
)

const RedisSlotCount int = 16384

type Cluster struct {
	nodes []*Node
}

type ClusterStats struct {
	nodesTotal int
	nodesUp    int
}

type Shard struct {
	MasterAddress   NodeAddress
	SlavesAddresses []NodeAddress
	FromSlot        int
	ToSlot          int
	masterIndex     int
	slaveIndices    []int
}

func NewCluster(baseDir string, conf *ClusterConf, binaries *Binaries) *Cluster {

	nodes := make([]*Node, len(conf.ListenPorts))

	for i, port := range conf.ListenPorts {
		nodes[i] = NewNode(baseDir, port, conf, binaries)
	}

	return &Cluster{
		nodes: nodes,
	}
}

func (self *Cluster) CreateNodes() {
	for _, node := range self.nodes {
		node.Create()
	}
}

func (self *Cluster) Nodes() []*Node {
	result := make([]*Node, len(self.nodes))
	copy(result, self.nodes)

	return result
}

type NodeWithState struct {
	Node *Node
	IsUp bool
}

type NodeWithStateByState []NodeWithState

func (self NodeWithStateByState) Len() int      { return len(self) }
func (self NodeWithStateByState) Swap(i, j int) { self[i], self[j] = self[j], self[i] }
func (self NodeWithStateByState) Less(i, j int) bool {
	if self[i].IsUp == self[j].IsUp {
		return false
	} else {
		return self[i].IsUp
	}
}

func (self *Cluster) NodesByState() ([]*Node, int, error) {

	var origNodes = self.Nodes()
	var nodesWithStates = make([]NodeWithState, len(origNodes))

	for i, node := range origNodes {

		if isUp, err := node.IsUp(); err != nil {
			return nil, -1, err
		} else {
			nodesWithStates[i] = NodeWithState{Node: node, IsUp: isUp}
		}
	}

	sort.Sort(NodeWithStateByState(nodesWithStates))

	var splitIndex = sort.Search(len(nodesWithStates), func(i int) bool {
		return !nodesWithStates[i].IsUp
	})

	var nodes = make([]*Node, len(nodesWithStates))

	for i, nodeWithState := range nodesWithStates {
		nodes[i] = nodeWithState.Node
	}

	return nodes, splitIndex, nil
}

func (self *Cluster) NodesCount() int {
	return len(self.nodes)
}

func (self *Cluster) Start() error {
	var err error = nil

	for _, node := range self.nodes {
		err = node.Start()
	}

	return err
}

func (self *Cluster) Stop() error {
	var err error = nil

	for _, node := range self.nodes {
		err = node.Stop()
	}

	return err
}

func (self *Cluster) Kill() error {
	var err error = nil

	for _, node := range self.nodes {
		err = node.Kill()
	}

	return err
}

func (self *Cluster) RandomNode(isUp bool) (*Node, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if nodes, splitIndex, err := self.NodesByState(); err != nil {
		return nil, err
	} else if splitIndex < 1 {
		return nil, ClusterIsDownError
	} else {
		return nodes[r.Intn(splitIndex)], nil
	}
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

func (self *Cluster) PrepareSlotDistribution(replicas int) []Shard {
	if replicas < 0 {
		panic("Number of replicas should be greater than zero")
	}

	nodesWithDataCount := replicas + 1

	nodesCount := len(self.nodes)

	if nodesWithDataCount > nodesCount {
		panic("Number of replicas should be less then nodes count")
	}

	mastersCount := nodesCount / nodesWithDataCount

	result := make([]Shard, mastersCount)

	slotsPerShard := RedisSlotCount / mastersCount

	for i, node := range self.nodes[:mastersCount] {
		result[i].MasterAddress = node.Address()
		result[i].masterIndex = i

		fromSlot := slotsPerShard * i

		result[i].FromSlot = fromSlot
		result[i].ToSlot = fromSlot + slotsPerShard
	}

	result[mastersCount-1].ToSlot = RedisSlotCount

	for j, node := range self.nodes[mastersCount:] {
		i := j % mastersCount
		result[i].SlavesAddresses = append(result[i].SlavesAddresses, node.Address())
		result[i].slaveIndices = append(result[i].slaveIndices, mastersCount+j)
	}

	return result
}

func (self *Cluster) ApplySlotDistribution(shards []Shard) error {
	firstNode := self.nodes[shards[0].masterIndex]

	for _, shard := range shards[1:] {
		meetErr := firstNode.ClusterMeet(shard.MasterAddress).Run()

		if meetErr != nil {
			return meetErr
		}
	}

	for _, shard := range shards {
		masterNode := self.nodes[shard.masterIndex]
		masterNodeId, masterNodeIdErr := masterNode.Id()

		if masterNodeIdErr != nil {
			return masterNodeIdErr
		}

		addSlotsErr := masterNode.ClusterAddSlots(shard.FromSlot, shard.ToSlot).Run()

		if addSlotsErr != nil {
			return addSlotsErr
		}

		for _, slaveIndex := range shard.slaveIndices {
			slaveNode := self.nodes[slaveIndex]

			clusterMeetErr := masterNode.ClusterMeet(slaveNode.Address()).Run()

			if clusterMeetErr != nil {
				return clusterMeetErr
			}

			clusterReplicateErr := slaveNode.ClusterReplicate(masterNodeId).Run()

			if clusterReplicateErr != nil {
				return clusterReplicateErr
			}
		}
	}

	return nil
}
