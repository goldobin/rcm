package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"os"
)

const MaxTcpPort = 65535
const RedisGossipPortIncrement = 10000

var bold = color.New(color.Bold).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

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

func main() {

	clusterSet, err := NewClusterSetAtHomeDir()

	if err != nil {
		failureCausedByError(err)
	}

	app := cli.NewApp()

	app.Name = "Redis Cluster Manager"
	app.Usage = ""
	app.Version = "0.0.1"

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
				validate(!clusterSet.Exists(name), "Cluster with %s already exists", bold(name))

				nodeCount := c.Int("nodes")
				maxPort := MaxTcpPort - RedisGossipPortIncrement - nodeCount
				startPort := c.Int("start-port")

				validate(nodeCount > 1, "Cluster should have at least 2 nodes")

				validate(
					startPort > 0 && startPort < maxPort,
					"Start port out of range of allowed ports (1-%s)",
					maxPort)

				echo("Creating cluster %s...", bold(name))

				ports := make([]int, nodeCount)

				for i := 0; i < nodeCount; i++ {
					ports[i] = startPort + i
				}

				clusterSet.Create(name, ClusterConf{
					ListenHost:  c.String("listen"),
					Ports:       ports,
					Persistence: c.Bool("persistance"),
				})
			},
		},
		cli.Command{
			Name:  "remove",
			Usage: "Removes existing cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster with %s does not exists", bold(name))

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
				validate(clusterSet.Exists(name), "Cluster with %s does not exists", bold(name))

				cluster, _ := clusterSet.Open(name)
				cluster.Start()
			},
		},
		cli.Command{
			Name:  "stop",
			Usage: "Stops the cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster with %s does not exists", bold(name))

				cluster, _ := clusterSet.Open(name)
				cluster.Stop()
			},
		},
		cli.Command{
			Name:  "list",
			Usage: "Lists available clusters",
			Action: func(c *cli.Context) {

				names := clusterSet.ListNames()

				for _, name := range names {
					echo(name)
				}
			},
		},
		cli.Command{
			Name:  "nodes",
			Usage: "Lists nodes in cluster",
		},
		cli.Command{
			Name:  "cli",
			Usage: "Opens a redis-cli session with random node",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				validate(clusterSet.Exists(name), "Cluster with %s does not exists", bold(name))

				cluster, _ := clusterSet.Open(name)
				cluster.Cli()
			},
		},
	}

	app.Run(os.Args)
}
