package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"os"
)

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

func validate(condition bool, errorMessage string) {

	if condition {
		return
	}

	fmt.Printf("%s %s\n", red("ERROR"), errorMessage)
	os.Exit(0)
}

func echo(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func main() {

	app := cli.NewApp()

	app.Name = "rcm"
	app.Usage = "Redis Cluster Manager"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "create",
			Usage: "Create a new cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")
				echo("Creating cluster %s...\n", bold(name))
			},
		},
		cli.Command{
			Name:  "remove",
			Usage: "Remove existing cluster",
			Action: func(c *cli.Context) {
				name := first(c.Args())

				validate(len(name) > 0, "Name of the cluster is required")

				if ask("Do you really want to remove cluster %s", bold(name)) {
					echo("Removing cluster %s\n", bold(name))
				} else {
					echo("Aborted.\n")
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
