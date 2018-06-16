package main

import "transfer.sh/cmd"

func main() {
	app := cmd.New()
	app.RunAndExitOnError()
}
