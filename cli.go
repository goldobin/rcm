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

	fatal(format, args...)
}

func echo(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func fatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	fmt.Printf("%s %s\n", red("ERROR"), message)
	os.Exit(0)
}

func fatalError(err error) {
	fmt.Printf("%s %s\n", red("ERROR"), err)
	os.Exit(0)
}

func main() {

	clusters, err := NewClustersInHomeDir()

	if err != nil {
		fatalError(err)
	}

	app := cli.NewApp()

	app.Name = "rcm"
	app.Usage = "Redis Cluster Manager"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "create",
			Usage: "Create a new cluster",
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
				validate(!clusters.Exists(name), "Cluster with %s already exists", bold(name))

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

				clusters.New(name, ClusterConf{
					ListenHost:  c.String("listen"),
					Ports:       ports,
					Persistence: c.Bool("persistance"),
				})
			},
		},
		cli.Command{
			Name:  "remove",
			Usage: "Remove existing cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")

				if ask("Do you really want to remove cluster %s", bold(name)) {
					echo("Removing cluster %s", bold(name))
				} else {
					echo("Aborted.")
				}
			},
		},
		cli.Command{
			Name:  "start",
			Usage: "Start the cluster",
		},
		cli.Command{
			Name:  "stop",
			Usage: "Stop the cluster",
		},
		cli.Command{
			Name:  "list",
			Usage: "List available clusters",
			Action: func(c *cli.Context) {

				names := clusters.ListNames()

				for _, name := range names {
					echo(name)
				}
			},
		},
		cli.Command{
			Name:  "nodes",
			Usage: "List nodes in cluster",
		},
		cli.Command{
			Name:  "cli",
			Usage: "Open a redis-cli session with random node",
		},
	}

	app.Run(os.Args)
}
