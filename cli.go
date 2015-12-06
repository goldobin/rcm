package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"os"
	"os/user"
	"path"
	"regexp"
	"sort"
	"strings"
)

const MaxTcpPort = 65535
const RedisGossipPortIncrement = 10000
const MaxClusterNameDisplayLength = 32
const RcmHome string = ".rcm"

var bold = color.New(color.Bold).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()

func first(args []string) string {
	if len(args) > 0 {
		return args[0]
	} else {
		return ""
	}
}

func ask(format string, args ...interface{}) bool {

	fmt.Printf(format+" "+bold("y/N:"), args...)

	var answer string
	fmt.Scanf("%s", &answer)

	return answer == "y"
}

func validate(condition bool, format string, args ...interface{}) {

	if condition {
		return
	}

	failure(format, args...)
}

func echo(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	fmt.Printf("%s %s\n", green("OK "), message)
	os.Exit(0)
}

func failure(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	fmt.Printf("%s %s\n", red("ERR"), message)
	os.Exit(0)
}

func failureCausedByError(err error) {
	fmt.Printf("%s %s\n", red("ERR"), err)
	os.Exit(0)
}

func shorter(name string, maxDisplayLength int) string {
	maxLengthWithoutCommas := maxDisplayLength - 3

	if len(name) < maxLengthWithoutCommas {
		return name
	} else {
		return name[0:maxLengthWithoutCommas] + "..."
	}
}

func main() {

	clusterNameRegEx := regexp.MustCompile(`^[\w+\-\.]+$`)

	usr, err := user.Current()
	if err != nil {
		failure("Can't determine user's home directory")
		return
	}

	binaries, err := NewBinaries()

	if err != nil {
		failureCausedByError(err)
	}

	clusterSet, err := NewClusterSet(path.Join(usr.HomeDir, RcmHome), binaries)

	if err != nil {
		failureCausedByError(err)
	}

	app := cli.NewApp()

	app.Name = "RCM"
	app.Usage = "Redis Cluster Manager"
	app.Version = "0.0.1"
	app.Copyright = "2015, Oleksandr Goldobin"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Oleksandr Goldobin",
			Email: "alex.goldobin@gmail.com",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "create",
			Usage: "Creates a new cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "listen, l",
					Value: "127.0.0.1",
					Usage: "listen on host for incomming connection (host to bind to)",
				},
				cli.IntFlag{
					Name:  "nodes, n",
					Value: 6,
					Usage: "number of nodes to create",
				},
				cli.BoolFlag{
					Name:  "persistance, s",
					Usage: "enable persistance",
				},
				cli.IntFlag{
					Name:  "start-port, p",
					Value: 10001,
					Usage: "port of the first node",
				},
			},
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(
					clusterNameRegEx.MatchString(name),
					"Illegal cluster name. The name should match %v",
					clusterNameRegEx)
				validate(!clusterSet.Exists(name), "Cluster %s already exists", bold(name))

				nodeCount := c.Int("nodes")
				maxPort := MaxTcpPort - RedisGossipPortIncrement - nodeCount
				startPort := c.Int("start-port")

				validate(nodeCount > 1, "Cluster should have at least 2 nodes")

				validate(
					startPort > 0 && startPort < maxPort,
					"Start port out of range of allowed ports (1-%s)",
					maxPort)

				listenIp := c.String("listen")

				ports := make([]int, nodeCount)
				for i := 0; i < nodeCount; i++ {
					ports[i] = startPort + i
				}

				if ask(
					"Create clustrer %s with %v nodes listening on %v:%v?",
					bold(name),
					nodeCount,
					listenIp,
					ports) {
					echo("Creating cluster %s...", bold(name))

					clusterSet.Create(
						name,
						&ClusterConf{
							ListenIp:    listenIp,
							ListenPorts: ports,
							Persistence: c.Bool("persistance"),
						})
				} else {
					echo("Aborting.")
				}
			},
		},
		cli.Command{
			Name:    "remove",
			Aliases: []string{"rm"},
			Usage:   "Removes existing cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if ask("Do you really want to remove cluster %s?", bold(name)) {
					echo("Removing cluster %s...", bold(name))

					err := clusterSet.Remove(name)

					if err != nil {
						failure("Can't remove cluster %s", bold(name))
					} else {
						success("Cluster %s has been successfully removed", bold(name))
					}
				} else {
					echo("Aborted.")
				}
			},
		},
		cli.Command{
			Name:  "start",
			Usage: "Starts the cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if cluster, err := clusterSet.Open(name); err != nil {
					failureCausedByError(err)
				} else if err := cluster.Start(); err != nil {
					failureCausedByError(err)
				}
			},
		},
		cli.Command{
			Name:  "stop",
			Usage: "Stops the cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if cluster, err := clusterSet.Open(name); err != nil {
					failureCausedByError(err)
				} else if err := cluster.Stop(); err != nil {
					failureCausedByError(err)
				}
			},
		},
		cli.Command{
			Name:  "distribute-slots",
			Usage: "Distributes slots in cluster",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "replicas, r",
					Value: 1,
					Usage: "number of data replicas",
				},
			},
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				cluster, err := clusterSet.Open(name)

				if err != nil {
					failureCausedByError(err)
				}

				replicas := c.Int("replicas")

				validate(
					replicas > 0 && replicas < cluster.NodesCount(),
					"Number of replicas should be greater then zero and less then node count(%v)", cluster.NodesCount())

				shards := cluster.PrepareSlotDistribution(replicas)

				for _, shard := range shards {
					slotRange := fmt.Sprintf("%v-%v", shard.FromSlot, shard.ToSlot-1)

					slaves := make([]string, len(shard.SlavesAddresses))

					for i, slaveAddress := range shard.SlavesAddresses {
						slaves[i] = slaveAddress.String()
					}

					echo("%-11s %20s %v", slotRange, bold(shard.MasterAddress), strings.Join(slaves, " "))
				}

				if ask("Do you want to proceed?") {
					cluster.ApplySlotDistribution(shards)
				} else {
					echo("Aborted.")
				}
			},
		},
		cli.Command{
			Name:  "list",
			Usage: "Lists available clusters",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "short, s",
					Usage: "display only names of clusters",
				},
			},
			Action: func(c *cli.Context) {

				names, err := clusterSet.ListNames()

				if err != nil {
					failureCausedByError(err)
				}

				sort.Strings(names)

				isShort := c.Bool("short")

				if isShort {
					for _, name := range names {
						echo(name)
					}
					return
				}

				for _, name := range names {
					cluster, err := clusterSet.Open(name)

					if err != nil {
						echo("%-40s %s %s", bold(name), red("ERROR"), "Can't open cluster")
						continue
					}

					stats, err := cluster.Stats()

					if err != nil {
						echo("%-40s %s %s", bold(name), red("ERROR"), "Can't fetch cluster stats")
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

					echo("%-40s %s", bold(shorter(name, MaxClusterNameDisplayLength)), status)
				}
			},
		},
		cli.Command{
			Name:  "ps",
			Usage: "Lists cluster's processes",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "short, s",
					Usage: "display only pids of nodes",
				},
			},
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if cluster, err := clusterSet.Open(name); err != nil {
					failureCausedByError(err)
				} else {
					isShort := c.Bool("short")

					for _, node := range cluster.Nodes() {

						var pid int

						if p, err := node.Pid(); err != nil {
							pid = -1
						} else {
							pid = p
						}

						if isShort {
							echo("%v", pid)
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

						echo("%-5v %-20s %s", pid, node.Address(), state)
					}
				}
			},
		},
		cli.Command{
			Name:  "info",
			Usage: "Executes `cluter info` command at random node",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if cluster, err := clusterSet.Open(name); err != nil {
					failureCausedByError(err)
				} else if b, err := cluster.RandomNode().ClusterInfo().Output(); err != nil {
					failureCausedByError(err)
				} else if _, err := os.Stdout.Write(b); err != nil {
					failureCausedByError(err)
				}
			},
		},
		cli.Command{
			Name:  "nodes",
			Usage: "Executes `cluter nodes` command at random node",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if cluster, err := clusterSet.Open(name); err != nil {
					failureCausedByError(err)
				} else if b, err := cluster.RandomNode().ClusterNodes().Output(); err != nil {
					failureCausedByError(err)
				} else if _, err := os.Stdout.Write(b); err != nil {
					failureCausedByError(err)
				}
			},
		},
		cli.Command{
			Name:  "slots",
			Usage: "Executes `cluter slots` command at random node",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if cluster, err := clusterSet.Open(name); err != nil {
					failureCausedByError(err)
				} else if b, err := cluster.RandomNode().ClusterSlots().Output(); err != nil {
					failureCausedByError(err)
				} else if _, err := os.Stdout.Write(b); err != nil {
					failureCausedByError(err)
				}
			},
		},
		cli.Command{
			Name:  "cli",
			Usage: "Opens a redis-cli session with random cluster node",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster %s does not exists", bold(name))

				if cluster, err := clusterSet.Open(name); err != nil {
					failureCausedByError(err)
				} else {
					cluster.RandomNode().Cli(c.Args()[1:]...)
				}
			},
		},
	}

	app.Run(os.Args)
}
