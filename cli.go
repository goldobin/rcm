package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"os/user"
	"path"
)

const RcmHome string = ".rcm"

func first(args []string) string {
	if len(args) > 0 {
		return args[0]
	} else {
		return ""
	}
}

func printError(err interface{}) {
	if err != nil {
		fmt.Printf("%s %s\n", red("ERROR"), err)
	}
}

func main() {
	usr, err := user.Current()
	if err != nil {
		printError(err)
		return
	}

	binaries, err := NewBinaries()

	if err != nil {
		printError(err)
		return
	}

	clusterSet, err := NewClusterSet(path.Join(usr.HomeDir, RcmHome), binaries)

	if err != nil {
		printError(err)
		return
	}

	controller := NewController(NewConsoleView(), clusterSet)

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
					Value: 9001,
					Usage: "port of the first node",
				},
			},
			Action: func(c *cli.Context) {
				err := controller.Create(
					first(c.Args()),
					CreateProperties{
						nodesCount:                c.Int("nodes"),
						startPort:                 c.Int("start-port"),
						listenIp:                  c.String("listen"),
						persistence:               c.Bool("persistance"),
						performFinalConfiguration: false,
						sayYes: false,
					})
				printError(err)
			},
		},
		cli.Command{
			Name:    "remove",
			Aliases: []string{"rm"},
			Usage:   "Removes existing cluster",
			Action: func(c *cli.Context) {
				err := controller.Remove(first(c.Args()), false)
				printError(err)
			},
		},
		cli.Command{
			Name:  "start",
			Usage: "Starts the cluster",
			Action: func(c *cli.Context) {
				err := controller.Start(first(c.Args()))
				printError(err)
			},
		},
		cli.Command{
			Name:  "stop",
			Usage: "Stops the cluster",
			Action: func(c *cli.Context) {
				err := controller.Stop(first(c.Args()))
				printError(err)
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
				err := controller.DistributeSlots(first(c.Args()), c.Int("replicas"), false)
				printError(err)
			},
		},
		cli.Command{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "Lists available clusters",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "short, s",
					Usage: "display only names of clusters",
				},
			},
			Action: func(c *cli.Context) {
				err := controller.List(c.Bool("short"))
				printError(err)
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
				err := controller.Ps(first(c.Args()), c.Bool("short"))
				printError(err)
			},
		},
		cli.Command{
			Name:        "damage",
			Usage:       "Damage cluster to some degree",
			Description: "Randomily starts/stops nodes in cluster to achieve specified number of live nodes",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "nodes, n",
					Value: "66%",
					Usage: "the amount of nodes that should be up in cluster",
				},
			},
			Action: func(c *cli.Context) {
				err := controller.Damage(first(c.Args()), c.String("count"))
				printError(err)
			},
		},
		cli.Command{
			Name:  "info",
			Usage: "Executes `cluter info` command at random node",
			Action: func(c *cli.Context) {
				err := controller.Info(first(c.Args()))
				printError(err)
			},
		},
		cli.Command{
			Name:  "nodes",
			Usage: "Executes `cluter nodes` command at random node",
			Action: func(c *cli.Context) {
				err := controller.Nodes(first(c.Args()))
				printError(err)
			},
		},
		cli.Command{
			Name:  "slots",
			Usage: "Executes `cluter slots` command at random node",
			Action: func(c *cli.Context) {
				err := controller.Slots(first(c.Args()))
				printError(err)
			},
		},
		cli.Command{
			Name:  "cli",
			Usage: "Opens a redis-cli session with random cluster node",
			Action: func(c *cli.Context) {
				err := controller.Cli(first(c.Args()), c.Args()[1:])
				printError(err)
			},
		},
	}

	app.Run(os.Args)
}
