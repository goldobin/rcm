package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	MinClusterNameLength        = 1
	MaxClusterNameDisplayLength = 32
	MinNodesCount               = 2
	MinTcpPort                  = 1
	MaxTcpPort                  = 65535
	RedisGossipPortIncrement    = 10000
)

var (
	clusterNameRegEx = regexp.MustCompile(`^[\w+\-\.]+$`)

	ClusterNameRequiredError = errors.New("Name of the cluster is required")
	IllegalClusterNameError  = fmt.Errorf(
		"Illegal cluster name. The name should match %v",
		clusterNameRegEx)
	NodesCountRequiredError  = errors.New("Nodes count is required")
	IllegalPercentValueError = errors.New("Illegal percent value. Should be in rage 0.01..100")
	ClusterIsDownError       = errors.New("All cluster nodes are down")
)

func ClusterExistsError(clusterName string) error {
	return fmt.Errorf("Cluster %s already exists", clusterName)
}

func ClusterDoesNotExistError(clusterName string) error {
	return fmt.Errorf("Cluster %s does not exists", clusterName)
}

func IllegalReplicaCount(clusterNodeCount int) error {
	return fmt.Errorf("Number of replicas should be in range 0..%v", clusterNodeCount)
}

func TooFewNumberOfNodesError() error {
	return errors.New("Cluster should have at least 2 nodes")
}

func PortOutOfRangeError(maxPort int) error {
	return fmt.Errorf("Start port out of range of allowed ports (1-%v)", maxPort)
}

func IllegalNodeCount(availableNodeCount int) error {
	return fmt.Errorf("Node count should be in range 1..%v (up nodes)", availableNodeCount)
}

type CreateProperties struct {
	nodesCount                int
	listenIp                  string
	startPort                 int
	persistence               bool
	performFinalConfiguration bool
	sayYes                    bool
}

func NewController(view View, clusterSet *ClusterSet) *Controller {
	return &Controller{
		view:       view,
		clusterSet: clusterSet,
	}
}

type Controller struct {
	view       View
	clusterSet *ClusterSet
}

type clusterCmdSupplier func(*Cluster) (*exec.Cmd, error)

func (self *Controller) Create(clusterName string, props CreateProperties) error {

	if len(clusterName) < MinClusterNameLength {
		return ClusterNameRequiredError
	}

	if !clusterNameRegEx.MatchString(clusterName) {
		return IllegalClusterNameError
	}

	if self.clusterSet.Exists(clusterName) {
		return ClusterExistsError(clusterName)
	}

	maxPort := MaxTcpPort - RedisGossipPortIncrement - props.nodesCount
	if props.nodesCount < MinNodesCount {
		return TooFewNumberOfNodesError()
	}

	if props.startPort < MinTcpPort || props.startPort > maxPort-1 {
		return PortOutOfRangeError(maxPort)
	}

	ports := make([]int, props.nodesCount)
	for i := 0; i < props.nodesCount; i++ {
		ports[i] = props.startPort + i
	}

	if self.view.Ask(
		"Create clustrer %s with %v nodes listening on %v:%v?",
		bold(clusterName),
		props.nodesCount,
		props.listenIp,
		ports) {

		self.view.Echo("Creating cluster %s...", bold(clusterName))

		_, err := self.clusterSet.Create(
			clusterName,
			&ClusterConf{
				ListenIp:    props.listenIp,
				ListenPorts: ports,
				Persistence: props.persistence,
			})

		if err != nil {
			return err
		} else {
			self.view.Success(
				"Cluster nodes created. To complete cluster clreation 'start' and 'distribute-slots' " +
					"operations should be performed")
		}
	} else {
		self.view.Aborted()
	}

	return nil
}

func (self *Controller) Remove(clusterName string, sayYes bool) error {
	if len(clusterName) < MinClusterNameLength {
		return ClusterNameRequiredError
	}

	if !self.clusterSet.Exists(clusterName) {
		return ClusterDoesNotExistError(clusterName)
	}

	if self.view.Ask("Do you really want to remove cluster %s?", bold(clusterName)) {
		self.view.Echo("Removing cluster %s...", bold(clusterName))

		err := self.clusterSet.Remove(clusterName)

		if err != nil {
			return err
		} else {
			self.view.Success("Cluster %s has been successfully removed", bold(clusterName))
		}
	} else {
		self.view.Echo("Aborted.")
	}

	return nil
}

func (self *Controller) Start(clusterName string) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else {
		return cluster.Start()
	}
}

func (self *Controller) Stop(clusterName string) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else {
		return cluster.Stop()
	}
}

func (self *Controller) DistributeSlots(clusterName string, replicas int, sayYes bool) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else {
		if replicas < 1 && replicas < cluster.NodesCount() {
			return IllegalReplicaCount(cluster.NodesCount())
		}

		shards := cluster.PrepareSlotDistribution(replicas)

		for _, shard := range shards {
			slotRange := fmt.Sprintf("%v-%v", shard.FromSlot, shard.ToSlot-1)

			slaves := make([]string, len(shard.SlavesAddresses))

			for i, slaveAddress := range shard.SlavesAddresses {
				slaves[i] = slaveAddress.String()
			}

			self.view.Echo("%-11s %20s %v", slotRange, bold(shard.MasterAddress), strings.Join(slaves, " "))
		}

		if self.view.Ask("Do you want to proceed?") {
			cluster.ApplySlotDistribution(shards)
		} else {
			self.view.Aborted()
		}

		return nil
	}
}

func (self *Controller) List(short bool) error {
	if names, err := self.clusterSet.ListNames(); err != nil {
		return err
	} else {
		sort.Strings(names)

		for _, name := range names {
			if short {
				self.view.Echo(name)
			} else {
				cluster, err := self.clusterSet.Open(name)

				if err != nil {
					self.view.Echo("%-40s %s %s", bold(name), red("ERROR"), "Can't open cluster")
					continue
				}

				stats, err := cluster.Stats()

				if err != nil {
					self.view.Echo("%-40s %s %s", bold(name), red("ERROR"), "Can't fetch cluster stats")
					continue
				}

				nodeUpRatio := fmt.Sprintf("(%v/%v)", stats.nodesUp, stats.nodesTotal)

				var status string

				if stats.nodesUp == 0 {
					status = yellow("DOWN" + nodeUpRatio)
				} else if stats.nodesUp < stats.nodesTotal {
					status = cyan("PARTIALLY UP" + nodeUpRatio)
				} else {
					status = green("UP" + nodeUpRatio)
				}

				self.view.Echo("%-40s %s", bold(shorter(name, MaxClusterNameDisplayLength)), status)
			}
		}

		return nil
	}
}

func (self *Controller) Ps(clusterName string, short bool) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else {
		for _, node := range cluster.Nodes() {

			var pid int

			if p, err := node.Pid(); err != nil {
				pid = -1
			} else {
				pid = p
			}

			if short {
				self.view.Echo("%v", pid)
				continue
			}

			var state string

			if err != nil {
				state = red("ERROR")
			} else {
				if pid > 0 {
					state = green("UP")
				} else {
					state = yellow("DOWN")
				}
			}

			self.view.Echo("%-5v %-20s %s", pid, node.Address(), state)
		}
		return nil
	}
}

func determineNodesToStop(cluster *Cluster, str string) ([]*Node, error) {
	str = strings.TrimSpace(str)

	strLength := len(str)

	if strLength < 1 {
		return nil, NodesCountRequiredError
	}

	isPercent := str[strLength-1] == '%'

	if isPercent && strLength < 2 {
		return nil, NodesCountRequiredError
	}

	if upNodes, err := cluster.NodesByState(true); err != nil {
		return nil, err
	} else if upNodesCount := len(upNodes); upNodesCount < 1 {
		return nil, ClusterIsDownError
	} else {
		var nodesToStopCount int

		if isPercent {
			if percent, err := strconv.ParseFloat(str[:strLength-1], 32); err != nil || percent < 0.01 || percent > 100 {
				return nil, IllegalPercentValueError
			} else {
				percentOfNodesToStop := int(math.Ceil(float64(cluster.NodesCount()) * percent / 100.0))
				nodesToStopCount = percentOfNodesToStop - (cluster.NodesCount() - upNodesCount)

				if nodesToStopCount < 0 {
					nodesToStopCount = 0
				}
			}
		} else {
			if count, err := strconv.Atoi(str); err != nil || count < 1 || count > upNodesCount {
				return nil, IllegalNodeCount(upNodesCount)
			} else {
				nodesToStopCount = count
			}
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		nodeIndexesToStop := make(map[int]bool)

		for len(nodeIndexesToStop) < nodesToStopCount {
			nodeIndexesToStop[r.Intn(upNodesCount)] = true
		}

		result := upNodes[:0]

		for i, _ := range nodeIndexesToStop {
			result = append(result, upNodes[i])
		}

		return result, nil
	}
}

func (self *Controller) Damage(clusterName string, nodesCountStr string) error {

	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else if nodesToStop, err := determineNodesToStop(cluster, nodesCountStr); err != nil {
		return err
	} else if len(nodesToStop) < 1 {
		self.view.Echo("Nothing to stop. Cluster is down or already damaged to specified degree")
	} else if self.view.Ask("Will stop %v nodes. Proceed?", len(nodesToStop)) {
		self.view.Echo("Stopping nodes...")

		for _, node := range nodesToStop {
			if err := node.Stop(); err != nil {
				return err
			}
		}

		self.view.Success("Stopped %v nodes", len(nodesToStop))
	}

	return nil
}

func (self *Controller) Repair(clusterName string) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else if downNodes, err := cluster.NodesByState(false); err != nil {
		return err
	} else if self.view.Ask("Will start %v nodes. Proceed?", len(downNodes)) {
		for _, node := range downNodes {
			if err := node.Start(); err != nil {
				return err
			}
		}

		return nil
	} else {
		return nil
	}
}

func (self *Controller) Info(clusterName string) error {
	return self.execClusterCmd(clusterName, func(cluster *Cluster) (*exec.Cmd, error) {
		if node, err := cluster.RandomNode(true); err != nil {
			return nil, err
		} else {
			return node.ClusterInfo(), nil
		}
	})
}

func (self *Controller) Nodes(clusterName string) error {
	return self.execClusterCmd(clusterName, func(cluster *Cluster) (*exec.Cmd, error) {
		if node, err := cluster.RandomNode(true); err != nil {
			return nil, err
		} else {
			return node.ClusterNodes(), nil
		}
	})
}

func (self *Controller) Slots(clusterName string) error {
	return self.execClusterCmd(clusterName, func(cluster *Cluster) (*exec.Cmd, error) {
		if node, err := cluster.RandomNode(true); err != nil {
			return nil, err
		} else {
			return node.ClusterSlots(), nil
		}
	})
}

func (self *Controller) Cli(clusterName string, args []string) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else if node, err := cluster.RandomNode(true); err != nil {
		return err
	} else {
		return node.Cli(args...)
	}
}

func (self *Controller) openCluster(clusterName string) (*Cluster, error) {
	if len(clusterName) < MinClusterNameLength {
		return nil, ClusterNameRequiredError
	}

	if !self.clusterSet.Exists(clusterName) {
		return nil, ClusterDoesNotExistError(clusterName)
	}

	return self.clusterSet.Open(clusterName)
}

func (self *Controller) execClusterCmd(clusterName string, f clusterCmdSupplier) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else if cmd, err := f(cluster); err != nil {
		return err
	} else if b, err := cmd.Output(); err != nil {
		return err
	} else {
		_, err := os.Stdout.Write(b)
		return err
	}
}

func shorter(name string, maxDisplayLength int) string {
	maxLengthWithoutCommas := maxDisplayLength - 3

	if len(name) < maxLengthWithoutCommas {
		return name
	} else {
		return name[0:maxLengthWithoutCommas] + "..."
	}
}
