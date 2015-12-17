package main

import (
	"fmt"
)

type ConsoleView struct{}

func NewConsoleView() *ConsoleView {
	return &ConsoleView{}
}

func (self *ConsoleView) Ask(format string, args ...interface{}) bool {

	fmt.Printf(format+" "+bold("y/N:"), args...)

	var answer string
	fmt.Scanf("%s", &answer)

	return answer == "y"
}

func (self *ConsoleView) Aborted() {
	fmt.Printf("%s\n", yellow("Aborted."))
}

func (self *ConsoleView) Echo(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (self *ConsoleView) Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	fmt.Printf("%s %s\n", green("SUCCESS"), message)
}
