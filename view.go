package main

type View interface {
	Ask(format string, args ...interface{}) bool

	Aborted()

	Echo(format string, args ...interface{})

	Success(format string, args ...interface{})
}
