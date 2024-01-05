package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anthdm/run/pkg/runtime"
	"github.com/anthdm/run/pkg/spidermonkey"
	"github.com/anthdm/run/pkg/util"
	"github.com/anthdm/run/proto"
	"github.com/bananabytelabs/wazero"
	prot "google.golang.org/protobuf/proto"
)

func main() {
	b, err := os.ReadFile("examples/js/index.js")
	if err != nil {
		log.Fatal(err)
	}

	req := &proto.HTTPRequest{
		Method: "GET",
		URL:    "/login",
	}
	reqb, _ := prot.Marshal(req)

	cache := wazero.NewCompilationCache()
	in := bytes.NewReader(reqb)
	out := &bytes.Buffer{}
	args := runtime.InvokeArgs{
		Blob:  spidermonkey.WasmBlob,
		Env:   map[string]string{"FOO": "this is the FOO env"},
		In:    in,
		Out:   out,
		Cache: cache,
		Debug: true,
		Args:  []string{"", "-e", string(b)},
	}

	fmt.Println("==========================================================")
	fmt.Println("WITHOUT MOD CACHE")
	runtime.Invoke(context.Background(), args)

	fmt.Println("==========================================================")
	fmt.Println("WITH MOD CACHE")
	runtime.Invoke(context.Background(), args)
	fmt.Println("==========================================================")

	resp, status, err := util.ParseRuntimeHTTPResponse(out.String())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
	fmt.Println(status)
}
