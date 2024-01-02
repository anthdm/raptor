package main

import (
	"flag"
	"fmt"
	"os"
)

func printUsage() {
	fmt.Printf(`
Usage: hailstorm [OPTIONS] COMMAND

Run any application in the cloud and on the edge

Commands:
endpoint    Create a new endpoint	
`)
}

var ()

func main() {
	flagset := flag.NewFlagSet("cli", flag.ExitOnError)
	flagset.Parse(os.Args[1:])

	args := flagset.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	switch args[0] {
	case "endpoint":
		handleCreateEndpoint(args[1:])
	case "deploy":
		handleDeploy(args[1:])
	}
}

func handleCreateEndpoint(args []string) {
	fmt.Println("creating endpoint: ", args)
}

func handleDeploy(args []string) {
	fmt.Println("deploying: ", args)
}
