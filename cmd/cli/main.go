package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/anthdm/ffaas/pkg/api"
	"github.com/anthdm/ffaas/pkg/client"
	"github.com/anthdm/ffaas/pkg/config"
	"github.com/google/uuid"
)

func printUsage() {
	fmt.Printf(`
Usage: hailstorm [OPTIONS] COMMAND

Run any application in the cloud and on the edge

Options:
--env			Set and environment variable [--env foo=bar]
--config		Set the configuration path [--config path/to/config.toml] 

Commands:
endpoint		Create a new endpoint [options... endpoint "myendpoint"]
deploy			Deploy an app to the cloud [deploy <endpointID path/to/app.wasm>]
help			Show usage

`)
}

type stringList []string

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func (l *stringList) String() string {
	return strings.Join(*l, ":")
}

var (
	env        stringList
	endpointID string
	configFile string
)

func main() {
	flagset := flag.NewFlagSet("cli", flag.ExitOnError)
	flagset.Usage = printUsage
	flagset.StringVar(&endpointID, "endpoint", "", "")
	flagset.StringVar(&configFile, "config", "config.toml", "")
	flagset.Var(&env, "env", "")
	flagset.Parse(os.Args[1:])

	if err := config.Parse(configFile); err != nil {
		printErrorAndExit(err)
	}

	args := flagset.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	c := client.New(client.NewConfig().WithURL(config.GetApiUrl()))
	command := command{
		client: c,
	}

	switch args[0] {
	case "endpoint":
		command.handleCreateEndpoint(args[1:])
	case "deploy":
		command.handleDeploy(args[1:])
	case "help":
		printUsage()
	default:
		printUsage()
	}
}

type command struct {
	client *client.Client
}

func (c command) handleCreateEndpoint(args []string) {
	if len(args) != 1 {
		printUsage()
		return
	}
	params := api.CreateEndpointParams{
		Name:        args[0],
		Environment: makeEnvMap(env),
	}
	endpoint, err := c.client.CreateEndpoint(params)
	if err != nil {
		printErrorAndExit(err)
	}
	b, err := json.MarshalIndent(endpoint, "", "    ")
	if err != nil {
		printErrorAndExit(err)
	}
	fmt.Println(string(b))
}

func (c command) handleDeploy(args []string) {
	if len(args) != 2 {
		printUsage()
		return
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		printErrorAndExit(fmt.Errorf("invalid endpoint id given: %s", args[0]))
	}
	b, err := os.ReadFile(args[1])
	if err != nil {
		printErrorAndExit(err)
	}
	deploy, err := c.client.CreateDeploy(id, bytes.NewReader(b), api.CreateDeployParams{})
	if err != nil {
		printErrorAndExit(err)
	}
	b, err = json.MarshalIndent(deploy, "", "    ")
	if err != nil {
		printErrorAndExit(err)
	}
	fmt.Println(string(b))
	fmt.Println()
	fmt.Printf("deploy is live on: %s/%s\n", config.GetWasmUrl(), deploy.EndpointID)
}

func makeEnvMap(list []string) map[string]string {
	m := make(map[string]string, len(list))
	for _, value := range list {
		parts := strings.Split(value, "=")
		if len(parts) != 2 {
			printErrorAndExit(fmt.Errorf("env arguments need to be in the format of --env foo=bar --env name=bob"))
		}
		m[parts[0]] = parts[1]
	}
	return m
}

func printErrorAndExit(err error) {
	fmt.Println()
	fmt.Println("Error:")
	fmt.Println(err)
	fmt.Println()
	os.Exit(1)
}
