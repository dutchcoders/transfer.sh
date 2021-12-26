package main

import (
	"log"
	"os"

	"github.com/dutchcoders/transfer.sh/cmd"
)

func main() {
	app := cmd.New()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
