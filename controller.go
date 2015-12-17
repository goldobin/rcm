package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
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
)

func ClusterExistsError(clusterName string) error {
	return fmt.Errorf("Cluster %s already exists", clusterName)
}

func ClusterDoesNotExistError(clusterName string) error {
	return fmt.Errorf("Cluster %s does not exists", clusterName)
}

func IllegalReplicaCount(clusterNodeCount int) error {
	return fmt.Errorf("Number of replicas should be greater then zero and less then node count(%v)", clusterNodeCount)
}

func TooFewNumberOfNodesError() error {
	return errors.New("Cluster should have at least 2 nodes")
}

func PortOutOfRangeError(maxPort int) error {
	return fmt.Errorf("Start port out of range of allowed ports (1-%v)", maxPort)
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

type clusterCmdSupplier func(*Cluster) *exec.Cmd

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

func (self *Controller) Info(clusterName string) error {
	return self.execClusterCmd(clusterName, func(cluster *Cluster) *exec.Cmd {
		return cluster.RandomNode().ClusterInfo()
	})
}

func (self *Controller) Nodes(clusterName string) error {
	return self.execClusterCmd(clusterName, func(cluster *Cluster) *exec.Cmd {
		return cluster.RandomNode().ClusterNodes()
	})
}

func (self *Controller) Slots(clusterName string) error {
	return self.execClusterCmd(clusterName, func(cluster *Cluster) *exec.Cmd {
		return cluster.RandomNode().ClusterSlots()
	})
}

func (self *Controller) Cli(clusterName string, args []string) error {
	if cluster, err := self.openCluster(clusterName); err != nil {
		return err
	} else {
		return cluster.RandomNode().Cli(args...)
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
	} else if b, err := f(cluster).Output(); err != nil {
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
