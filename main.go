package main

import "github.com/gufertum/transfer.sh/cmd"

func main() {
	app := cmd.New()
	app.RunAndExitOnError()
}
