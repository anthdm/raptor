package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/anthdm/raptor/pkg/api"
	"github.com/anthdm/raptor/pkg/client"
	"github.com/anthdm/raptor/pkg/config"
	"github.com/anthdm/raptor/pkg/types"
	"github.com/google/uuid"
)

func printUsage() {
	fmt.Printf(`
Usage: raptor [OPTIONS] COMMAND

Options:
--env			Set and environment variable [--env foo=bar]
--config		Set the configuration path [--config path/to/config.toml] 
--runtime		Set the runtime of your application [--runtime go or js]

Commands:
endpoint		Create a new endpoint [options... endpoint "myendpoint"]
test			Test your application [options... test "path/to/app.wasm(js)"]

Subcommands:
create 			Create a new endpoint [options... endpoint create "myendpoint"]
list			List current endpoints

deploy			Deploy an app to the cloud [deploy <endpointID path/to/app.wasm>]
help			Show usage

`)
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

var (
	env        stringList
	endpointID string
	configFile string
	runtime    string
	addr       string
)

func main() {
	flagset := flag.NewFlagSet("cli", flag.ExitOnError)
	flagset.Usage = printUsage
	flagset.StringVar(&endpointID, "endpoint", "", "")
	flagset.StringVar(&configFile, "config", "config.toml", "")
	flagset.StringVar(&addr, "addr", ":3000", "")
	flagset.StringVar(&runtime, "runtime", "", "")

	flagset.Var(&env, "env", "")
	flagset.Parse(os.Args[1:])

	if err := config.Parse(configFile); err != nil {
		printErrorAndExit(err)
	}

	args := flagset.Args()
	if len(args) == 0 {
		printUsage()
	}

	c := client.New(client.NewConfig().WithURL(config.GetApiUrl()))
	command := command{
		client: c,
	}

	switch args[0] {
	case "endpoint":
		if len(args) < 2 {
			printUsage()
		}
		switch args[1] {
		case "rollback":
			command.handleRollback(args)
		case "create":
			command.handleCreateEndpoint(args)
		case "list":
			command.handleListEndpoints(args)
		default:
			printUsage()
		}
	case "deploy":
		if len(args) < 2 {
			printUsage()
		}
		command.handleDeploy(args[1:])
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

// endpoint rollback <endpointID> <deployID>
func (c command) handleRollback(args []string) {
	if len(args) != 4 {
		printUsage()
	}
	endpointID, err := uuid.Parse(args[2])
	if err != nil {
		printErrorAndExit(err)
	}
	deployID, err := uuid.Parse(args[3])
	if err != nil {
		printErrorAndExit(err)
	}
	params := api.CreateRollbackParams{DeployID: deployID}
	resp, err := c.client.RollbackEndpoint(endpointID, params)
	if err != nil {
		printErrorAndExit(err)
	}
	b, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		printErrorAndExit(err)
	}
	fmt.Println(string(b))
}

func (c command) handleListEndpoints(args []string) {
	endpoints, err := c.client.ListEndpoints()
	if err != nil {
		printErrorAndExit(err)
	}
	b, err := json.MarshalIndent(endpoints, "", "    ")
	if err != nil {
		printErrorAndExit(err)
	}
	fmt.Println(string(b))
}

func (c command) handleCreateEndpoint(args []string) {
	if len(args) != 3 {
		printUsage()
	}
	if !types.ValidRuntime(runtime) {
		printUsage()
	}
	params := api.CreateEndpointParams{
		Runtime:     runtime,
		Name:        args[2],
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
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		printErrorAndExit(fmt.Errorf("invalid endpoint id given: %s", args[0]))
	}
	b, err := os.ReadFile(args[1])
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
	fmt.Printf("deploy is live on: %s/%s\n", config.GetWasmUrl(), deploy.EndpointID)
}

func (c command) handleServeEndpoint(args []string) {
	fmt.Println("TODO")
	// b, err := os.ReadFile(args[0])
	// if err != nil {
	// 	printErrorAndExit(err)
	// }

	// out := &bytes.Buffer{}

	// http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if r.URL.Path == "/favicon.ico" {
	// 		return
	// 	}
	// 	preq, err := shared.MakeProtoRequest(uuid.NewString(), r)
	// 	if err != nil {
	// 		w.Write([]byte(err.Error()))
	// 		return
	// 	}
	// 	fmt.Println(preq)
	// 	reqb, err := proto.Marshal(preq)
	// 	if err != nil {
	// 		w.Write([]byte(err.Error()))
	// 		return
	// 	}

	// 	invokeArgs := run.InvokeArgs{
	// 		Blob: b,
	// 		Out:  out,
	// 		In:   bytes.NewBuffer(reqb),
	// 	}

	// 	if err := run.Invoke(r.Context(), invokeArgs); err != nil {
	// 		w.Write([]byte(err.Error()))
	// 		return
	// 	}

	// 	resp, status, err := shared.ParseRuntimeHTTPResponse(out.String())
	// 	if err != nil {
	// 		w.Write([]byte(err.Error()))
	// 		return
	// 	}
	// 	w.WriteHeader(status)
	// 	w.Write([]byte(resp))
	// }))
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
