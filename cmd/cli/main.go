package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/anthdm/raptor/internal/api"
	"github.com/anthdm/raptor/internal/client"
	"github.com/anthdm/raptor/internal/config"
	"github.com/anthdm/raptor/internal/types"
	"github.com/anthdm/raptor/internal/version"
	"github.com/google/uuid"
)

func printUsage() {
	fmt.Printf(`
Raptor cli v%s

Usage: raptor COMMAND

Commands:
  endpoint			Create a new endpoint
  publish			Publish a deployment to an endpoint
  deploy			Create a new deployment
  list-deploy		Lists all deployments
  help				Show usage

`, version.Version)
	os.Exit(0)
}

type stringList []string

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func (l *stringList) String() string {
	return strings.Join(*l, ":")
}

func main() {
	flagset := flag.NewFlagSet("cli", flag.ExitOnError)

	var configFile string
	flagset.StringVar(&configFile, "config", "config.toml", "The location of your raptor config file")

	flagset.Usage = printUsage
	flagset.Parse(os.Args[1:])

	if err := config.Parse(configFile); err != nil {
		printErrorAndExit(err)
	}

	args := flagset.Args()
	if len(args) == 0 {
		printUsage()
	}

	c := client.New(client.NewConfig().WithURL(config.ApiUrl()))
	command := command{
		client: c,
	}

	switch args[0] {
	case "publish":
		command.handlePublish(args[1:])
	case "endpoint":
		command.handleEndpoint(args[1:])
	case "deploy":
		command.handleDeploy(args[1:])
	case "list-deploy":
		command.handleListDeploy(args[1:])
	case "serve":
		if len(args) < 2 {
			printUsage()
		}
		command.handleServeEndpoint(args[1:])
	case "help":
		printUsage()
	default:
		printUsage()
	}
}

type command struct {
	client *client.Client
}

func (c command) handlePublish(args []string) {
	flagset := flag.NewFlagSet("endpoint", flag.ExitOnError)

	var deployID string
	flagset.StringVar(&deployID, "deploy", "", "The id of the deployment that you want to publish LIVE")
	_ = flagset.Parse(args)

	id, err := uuid.Parse(deployID)
	if err != nil {
		printErrorAndExit(err)
	}

	params := api.PublishParams{DeploymentID: id}
	resp, err := c.client.Publish(params)
	if err != nil {
		printErrorAndExit(err)
	}
	b, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		printErrorAndExit(err)
	}
	fmt.Println(string(b))
}

func (c command) handleEndpoint(args []string) {
	flagset := flag.NewFlagSet("endpoint", flag.ExitOnError)

	var name string
	flagset.StringVar(&name, "name", "", "The name of your endpoint")
	var runtime string
	flagset.StringVar(&runtime, "runtime", "", "The runtime of your endpoint (go or js)")
	var env stringList
	flagset.Var(&env, "env", "Environment variables for this endpoint")
	_ = flagset.Parse(args)

	if len(runtime) == 0 {
		fmt.Println("please provide a valid runtime [--runtime go, --runtime js]")
		os.Exit(1)
	}
	if !types.ValidRuntime(runtime) {
		fmt.Printf("invalid runtime %s, only go and js are currently supported\n", runtime)
		os.Exit(1)
	}
	if len(name) == 0 {
		fmt.Println("The name of the endpoint is not provided. --name <name>")
		os.Exit(1)
	}
	params := api.CreateEndpointParams{
		Runtime:     runtime,
		Name:        name,
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
	flagset := flag.NewFlagSet("deploy", flag.ExitOnError)

	var endpointID string
	flagset.StringVar(&endpointID, "endpoint", "", "The id of the endpoint to where you want to deploy")
	var file string
	flagset.StringVar(&file, "file", "", "The file location of your code that you want to deploy")
	_ = flagset.Parse(args)

	id, err := uuid.Parse(endpointID)
	if err != nil {
		printErrorAndExit(fmt.Errorf("invalid endpoint id given: %s", args[0]))
	}
	b, err := os.ReadFile(file)
	if err != nil {
		printErrorAndExit(err)
	}
	deploy, err := c.client.CreateDeployment(id, bytes.NewReader(b), api.CreateDeploymentParams{})
	if err != nil {
		printErrorAndExit(err)
	}
	b, err = json.MarshalIndent(deploy, "", "    ")
	if err != nil {
		printErrorAndExit(err)
	}
	fmt.Println(string(b))
	fmt.Println()
	fmt.Printf("deploy preview: %s/preview/%s\n", config.IngressUrl(), deploy.ID)
}

func (c command) handleListDeploy(args []string) {
	flagSet := flag.NewFlagSet("list-deploy", flag.ExitOnError)
	_ = flagSet.Parse(args)

	endpoint, err := c.client.ListDeployments()
	if err != nil {
		printErrorAndExit(err)
	}
	b, err := json.MarshalIndent(endpoint, "", "    ")
	if err != nil {
		printErrorAndExit(err)
	}
	fmt.Println(string(b))
}

func (c command) handleServeEndpoint(args []string) {
	fmt.Println("TODO")
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
