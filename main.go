package main

import "github.com/dutchcoders/transfer.sh/cmd"

func main() {
	app := cmd.New()
	app.RunAndExitOnError()
}
